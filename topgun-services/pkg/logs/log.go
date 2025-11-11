package logs

import (
	"fmt"
	"log"
	"os"
	"time"

	"topgun-services/pkg/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LogFileWriter struct {
	LogPath      string
	FileName     string
	PrintConsole bool
}

func (c *LogFileWriter) Write(body []byte) (n int, err error) {
	if c.PrintConsole {
		fmt.Printf("%+s", body)
	}

	var logPath string
	var logFileName string

	if len(c.LogPath) == 0 {
		logPath = "./log"
	} else {
		logPath = c.LogPath
	}
	err = os.MkdirAll(logPath, 0755)
	if err != nil {
		return
	}

	if len(c.FileName) == 0 {
		logFileName = fmt.Sprintf("access-%s.log", time.Now().Format("2006-01-02"))
	} else {
		logFileName = c.FileName
	}
	logFile, err := os.OpenFile(fmt.Sprintf("%s/%s", logPath, logFileName), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		return
	}
	defer logFile.Close()

	return logFile.Write(body)
}

func LogMiddleware(db *gorm.DB, queue chan models.Log) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		var userUUID uuid.UUID
		userID := c.Locals("user_id")
		if userID == nil {
			userUUID = uuid.Nil
		} else {
			parsedUUID, errUid := uuid.Parse(userID.(string))
			if errUid != nil {
				log.Printf("failed to parse user_id to uuid: %v", errUid)
			}
			userUUID = parsedUUID
		}
		entry := models.Log{
			At:            start,
			Status:        c.Response().StatusCode(),
			IP:            c.IP(),
			Method:        c.Method(),
			Host:          c.Hostname(),
			URL:           c.OriginalURL(),
			UserAgent:     c.Get("User-Agent"),
			Referer:       c.Get("Referer"),
			Authorization: c.Get("Authorization"),
			UserID:        userUUID,
			BytesRecv:     len(c.Request().Body()),
			BytesSent:     len(c.Response().Body()),
			ErrorMsg: func() string {
				if err != nil {
					return err.Error()
				}
				return ""
			}(),
		}

		select {
		case queue <- entry:
		default:

		}
		return err
	}
}

func StartLogWorker(db *gorm.DB, queue chan models.Log, batchSize int, flushEvery time.Duration, stop <-chan struct{}) {
	batch := make([]models.Log, 0, batchSize)
	ticker := time.NewTicker(flushEvery)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		_ = db.Create(&batch).Error
		batch = batch[:0]
	}

	for {
		select {
		case e := <-queue:
			batch = append(batch, e)
			if len(batch) >= batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-stop:
			flush()
			return
		}
	}
}
