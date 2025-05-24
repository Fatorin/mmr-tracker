package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/fatorin/mmr-tracker/database"
	"github.com/fatorin/mmr-tracker/models"
	"github.com/fatorin/mmr-tracker/utils"
	"github.com/gin-gonic/gin"
)

func GetScores(c *gin.Context) {
	category := c.Query("category")
	server := c.Query("server")
	name := c.Query("name")
	sortBy := strings.ToLower(c.DefaultQuery("sort_by", "score"))
	sortOrder := strings.ToUpper(c.DefaultQuery("sort_order", "DESC"))

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "25"))
	if limit <= 0 {
		limit = 25
	}
	if limit > 50 {
		limit = 50
	}

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if offset < 0 {
		offset = 0
	}

	// 驗證排序參數
	if sortBy != "name" && sortBy != "score" {
		sortBy = "score"
	}
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}

	// 查詢條件
	conditions := []string{"1=1"}
	args := []any{}

	if category != "" {
		conditions = append(conditions, "category = ?")
		args = append(args, category)
	}
	if server != "" {
		conditions = append(conditions, "server = ?")
		args = append(args, server)
	}
	if name != "" {
		conditions = append(conditions, "name LIKE ?")
		args = append(args, "%"+name+"%")
	}

	whereClause := strings.Join(conditions, " AND ")

	query := `
		SELECT id, name, score
		FROM scores
		WHERE ` + whereClause + `
		ORDER BY ` + sortBy + ` ` + sortOrder + `
		LIMIT ? OFFSET ?
	`
	args = append(args, limit, offset)

	var scores []models.Score
	if err := database.DB.Select(&scores, query, args...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	countQuery := `SELECT COUNT(*) FROM scores WHERE ` + whereClause
	var total int
	if err := database.DB.Get(&total, countQuery, args[:len(args)-2]...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pages, currentPage, hasNext := utils.Paginate(total, limit, offset)

	c.JSON(http.StatusOK, models.PaginationResult[models.Score]{
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		Page:    currentPage,
		Pages:   pages,
		HasNext: hasNext,
		Data:    scores,
	})
}
