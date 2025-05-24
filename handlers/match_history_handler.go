package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/fatorin/mmr-tracker/database"
	"github.com/fatorin/mmr-tracker/models"
	"github.com/fatorin/mmr-tracker/utils"
	"github.com/gin-gonic/gin"
)

type playerInfo struct {
	UserName string  `db:"username"`
	Pid      int     `db:"pid"`
	Servant  *string `db:"servant"`
	Kills    *int    `db:"kills"`
	Deaths   *int    `db:"deaths"`
	Assists  *int    `db:"assists"`
	Level    *int    `db:"level"`
}

func GetMatchHistories(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 {
		limit = 1
	}
	if limit > 10 {
		limit = 10
	}
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if offset < 0 {
		offset = 0
	}

	var games []models.Game
	queryGame := `
		SELECT id, map, datetime, duration
		FROM games
		ORDER BY datetime DESC
		LIMIT ? OFFSET ?
	`
	if err := database.DB.Select(&games, queryGame, limit, offset); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var matchHistories []models.MatchHistory

	for _, game := range games {
		gameID := game.ID

		var playerInfos []playerInfo
		queryServants := `
			SELECT
				wp.name AS username,
				wp.pid,
				v_servant.value_string AS servant,
				v_kills.value_int AS kills,
				v_deaths.value_int AS deaths,
				v_assists.value_int AS assists,
				v_level.value_int AS level
			FROM w3mmdplayers wp
			LEFT JOIN w3mmdvars v_servant ON v_servant.gameid = wp.gameid AND v_servant.pid = wp.pid AND v_servant.varname = 'servant'
			LEFT JOIN w3mmdvars v_kills   ON v_kills.gameid   = wp.gameid AND v_kills.pid   = wp.pid AND v_kills.varname   = 'kills'
			LEFT JOIN w3mmdvars v_deaths  ON v_deaths.gameid  = wp.gameid AND v_deaths.pid  = wp.pid AND v_deaths.varname  = 'deaths'
			LEFT JOIN w3mmdvars v_assists ON v_assists.gameid = wp.gameid AND v_assists.pid = wp.pid AND v_assists.varname = 'assists'
			LEFT JOIN w3mmdvars v_level   ON v_level.gameid   = wp.gameid AND v_level.pid   = wp.pid AND v_level.varname   = 'level'
			WHERE wp.gameid = ?
		`
		if err := database.DB.Select(&playerInfos, queryServants, gameID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var teamInfos []*string
		queryTeams := `
			SELECT value_string
			FROM w3mmdvars
			WHERE gameid = ? AND varname = 'team_info'
		`
		if err := database.DB.Select(&teamInfos, queryTeams, gameID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		teams := analyse(teamInfos, playerInfos)

		matchHistories = append(matchHistories, models.MatchHistory{
			ID:       game.ID,
			Map:      game.Map,
			DateTime: game.DateTime,
			Duration: game.Duration,
			Teams:    teams,
		})
	}

	var total int
	if err := database.DB.Get(&total, `SELECT COUNT(*) FROM games`); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pages, currentPage, hasNext := utils.Paginate(total, limit, offset)

	c.JSON(http.StatusOK, models.PaginationResult[models.MatchHistory]{
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		Page:    currentPage,
		Pages:   pages,
		HasNext: hasNext,
		Data:    matchHistories,
	})
}

func analyse(teamInfos []*string, playerInfos []playerInfo) []models.Team {
	teamMap := make(map[int]*models.Team)

	for _, teamInfo := range teamInfos {
		index, teamName, score, err := teamInfoSplit(teamInfo)

		if err != nil {
			fmt.Println("An attempt to parse team data failed:", err)
			continue
		}

		_, exist := teamMap[index]
		if exist {
			fmt.Println("team index dumplicate:", index)
			return nil
		}

		teamMap[index] = &models.Team{
			Index: index,
			Name:  teamName,
			Score: score,
		}
	}

	for _, playerInfo := range playerInfos {
		if playerInfo.Servant == nil {
			return nil
		}

		index, name, err := playerInfoSplit(*playerInfo.Servant)

		if err != nil {
			fmt.Println("An attempt to parse servant data failed:", err)
			continue
		}

		team, ok := teamMap[index]
		if !ok {
			fmt.Println("not found team index:", index)
			continue
		}

		servant := models.Servant{
			UserName: playerInfo.UserName,
			Name:     name,
			Level:    convertPointInt(playerInfo.Level),
			Kills:    convertPointInt(playerInfo.Kills),
			Deaths:   convertPointInt(playerInfo.Deaths),
			Assists:  convertPointInt(playerInfo.Assists),
		}

		team.Servants = append(team.Servants, servant)
		teamMap[index] = team
	}

	teams := make([]models.Team, 0, len(teamMap))
	for _, teamPtr := range teamMap {
		teams = append(teams, *teamPtr)
	}
	return teams
}

// teamInfoSplit 解析 teamInfo 字串，格式應為 "索引:隊伍名稱:分數"，例如 "1:Team A:12"。
// 此函式會去除字串前後的雙引號後分割成三部分，並將第一與第三項轉換為整數，回傳索引、隊伍名稱與分數。
//
// 參數：
//
//	teamInfo - 格式為帶雙引號的字串，例如 `"1:Team A:12"`
//
// 回傳值：
//
//	int    - 隊伍索引
//	string - 隊伍名稱
//	int    - 隊伍分數
//	error  - 若格式不正確或索引/分數轉換失敗，則回傳錯誤
func teamInfoSplit(teamInfo *string) (int, string, int, error) {
	if teamInfo == nil {
		return 0, "", 0, fmt.Errorf("not team info is nil")
	}

	cleaned := strings.Trim(*teamInfo, `"`)
	parts := strings.Split(cleaned, ":")
	if len(parts) != 3 {
		return 0, "", 0, fmt.Errorf("length is invalid")
	}

	index, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", 0, fmt.Errorf("convert index failed: %v", err)
	}

	score, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, "", 0, fmt.Errorf("convert score failed: %v", err)
	}

	return index, parts[1], score, nil
}

// playerInfoSplit 解析 servant 字串，格式應為 "隊伍ID:玩家名稱"，例如 "1:Siegfried"。
// 會去除字串前後的雙引號後進行分割，回傳隊伍 ID（int）、玩家名稱（string）與錯誤（若格式不正確或 ID 無法轉換）。
//
// 參數：
//
//	servant - 格式為帶雙引號的字串，例如 `"1:Siegfried"`
//
// 回傳值：
//
//	int - 隊伍 ID
//	string - 玩家名稱
//	error - 若格式錯誤或轉換失敗則回傳錯誤
func playerInfoSplit(servant string) (int, string, error) {
	cleaned := strings.Trim(servant, `"`)
	parts := strings.Split(cleaned, ":")
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("length is invalid")
	}

	teamID, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", fmt.Errorf("convert id failed: %v", err)
	}

	return teamID, parts[1], nil
}

func convertPointInt(value *int) int {
	if value == nil {
		return -1
	}

	return *value
}
