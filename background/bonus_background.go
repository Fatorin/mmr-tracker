package bonus

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

type BonusRecord struct {
	PID      int    `db:"pid"`
	Name     string `db:"name"`
	Category string `db:"category"`
	Flag     string `db:"flag"`
	Servant  string `db:"servant"`
	Bonus    float64
}

// 查找尚未處理過的 GameIDs
func GetUnprocessedGameIDs(db *sqlx.DB) ([]int, error) {
	var gameIDs []int
	sql := `
		SELECT g.id
		FROM games g
		LEFT JOIN game_bonus_processed p ON g.id = p.gameid
		WHERE p.id IS NULL
		ORDER BY g.id
	`
	err := db.Select(&gameIDs, sql)
	return gameIDs, err
}

// 處理單一場次的加分邏輯
func ProcessGameBonus(db *sqlx.DB, gameID int) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 抓出符合 bonus 條件的 player
	var bonuses []BonusRecord
	rawSQL := `
		SELECT
		  p.pid,
		  p.name,
		  p.category,
		  p.flag,
		  v.value_string AS servant
		FROM w3mmdplayers p
		JOIN w3mmdvars v ON v.gameid = p.gameid AND v.pid = p.pid
		WHERE p.gameid = ?
		  AND v.varname = 'servant'
		  AND (v.value_string LIKE '%Heracles%' OR v.value_string LIKE '%Lancelot%')
		  AND p.flag IN ('winner', 'loser')
	`
	err = tx.Select(&bonuses, rawSQL, gameID)
	if err != nil {
		return err
	}

	// 抓 server 名稱
	var server string
	err = tx.Get(&server, `SELECT server FROM games WHERE id = ? LIMIT 1`, gameID)
	if err != nil {
		return err
	}

	for _, b := range bonuses {
		if b.Flag == "winner" {
			b.Bonus = 1.5
		} else {
			b.Bonus = 0.5
		}

		// 嘗試讀取目前分數
		var score float64
		err := tx.Get(&score, `
			SELECT score FROM scores
			WHERE category = ? AND name = ? AND server = ?
			LIMIT 1`, b.Category, b.Name, server)

		if err == sql.ErrNoRows {
			// 新增初始分數
			_, err = tx.Exec(`
				INSERT INTO scores (category, name, server, score)
				VALUES (?, ?, ?, GREATEST(1000 + ?, 0))
			`, b.Category, b.Name, server, b.Bonus)
			log.Printf("🔰 新增分數: %s (%s) +%.1f", b.Name, b.Flag, b.Bonus)
		} else if err == nil {
			// 更新舊有分數
			_, err = tx.Exec(`
				UPDATE scores SET score = GREATEST(score + ?, 0)
				WHERE category = ? AND name = ? AND server = ?
			`, b.Bonus, b.Category, b.Name, server)
			log.Printf("✨ 更新分數: %s (%s) +%.1f", b.Name, b.Flag, b.Bonus)
		}

		if err != nil {
			return fmt.Errorf("failed to apply bonus to %s: %v", b.Name, err)
		}
	}

	// 標記為已處理
	_, err = tx.Exec(`
		INSERT INTO game_bonus_processed (gameid, processed_at)
		VALUES (?, ?)
	`, gameID, time.Now())
	if err != nil {
		return err
	}

	log.Printf("✅ 完成處理 GameID %d (%d 位玩家)", gameID, len(bonuses))
	return tx.Commit()
}
