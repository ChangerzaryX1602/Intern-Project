package logs

import (
	"strconv"
	"strings"
	"time"

	"topgun-services/pkg/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type LogResponse struct {
	Data       []models.Log `json:"data"`
	Total      int64        `json:"total"`
	Page       int          `json:"page"`
	Limit      int          `json:"limit"`
	TotalPages int          `json:"total_pages"`
}

func GetLogsHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		page, _ := strconv.Atoi(c.Query("page", "1"))
		limit, _ := strconv.Atoi(c.Query("limit", "50"))
		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 1000 {
			limit = 50
		}
		offset := (page - 1) * limit

		query := db.Model(&models.Log{})

		if search := c.Query("search"); search != "" {
			query = query.Where("ip LIKE ? OR method LIKE ? OR url LIKE ? OR user_agent LIKE ? OR error_msg LIKE ?",
				"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
		}

		if status := c.Query("status"); status != "" {
			query = query.Where("status = ?", status)
		}

		if method := c.Query("method"); method != "" {
			query = query.Where("method = ?", strings.ToUpper(method))
		}

		if ip := c.Query("ip"); ip != "" {
			query = query.Where("ip LIKE ?", "%"+ip+"%")
		}

		if userID := c.Query("user_id"); userID != "" {
			query = query.Where("user_id = ?", userID)
		}

		if fromDate := c.Query("from_date"); fromDate != "" {
			if parsedDate, err := time.Parse("2006-01-02", fromDate); err == nil {
				query = query.Where("at >= ?", parsedDate)
			}
		}
		if toDate := c.Query("to_date"); toDate != "" {
			if parsedDate, err := time.Parse("2006-01-02", toDate); err == nil {
				query = query.Where("at <= ?", parsedDate.Add(24*time.Hour))
			}
		}

		if hasError := c.Query("has_error"); hasError != "" {
			switch hasError {
			case "true":
				query = query.Where("error_msg != ''")
			case "false":
				query = query.Where("error_msg = ''")
			}
		}

		var total int64
		query.Count(&total)

		var logs []models.Log
		err := query.Order("at DESC").
			Offset(offset).
			Limit(limit).
			Find(&logs).Error

		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to fetch logs",
			})
		}

		totalPages := int((total + int64(limit) - 1) / int64(limit))

		return c.JSON(LogResponse{
			Data:       logs,
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		})
	}
}

func GetLogStatsHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var stats struct {
			TotalLogs    int64 `json:"total_logs"`
			ErrorLogs    int64 `json:"error_logs"`
			TodayLogs    int64 `json:"today_logs"`
			UniqueIPs    int64 `json:"unique_ips"`
			UniqueUsers  int64 `json:"unique_users"`
			AvgBytesRecv int64 `json:"avg_bytes_recv"`
			AvgBytesSent int64 `json:"avg_bytes_sent"`
		}

		db.Model(&models.Log{}).Count(&stats.TotalLogs)

		db.Model(&models.Log{}).Where("error_msg != ''").Count(&stats.ErrorLogs)

		today := time.Now().Format("2006-01-02")
		db.Model(&models.Log{}).Where("DATE(at) = ?", today).Count(&stats.TodayLogs)

		db.Model(&models.Log{}).Distinct("ip").Count(&stats.UniqueIPs)

		db.Model(&models.Log{}).Where("user_id != '00000000-0000-0000-0000-000000000000'").Distinct("user_id").Count(&stats.UniqueUsers)

		var avgResult struct {
			AvgRecv int64
			AvgSent int64
		}
		db.Model(&models.Log{}).Select("AVG(bytes_recv) as avg_recv, AVG(bytes_sent) as avg_sent").Scan(&avgResult)
		stats.AvgBytesRecv = avgResult.AvgRecv
		stats.AvgBytesSent = avgResult.AvgSent

		return c.JSON(stats)
	}
}
