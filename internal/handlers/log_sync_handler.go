package handlers

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"andorralee/pkg/utils"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// LogSyncRequest represents node -> central log push payload.
// type: heralding_auth | heralding_session | cowrie | mysql_honeypot | dionaea
// headers: X-SYNC-TOKEN or Authorization: Bearer <token>
type LogSyncRequest struct {
	SourceHost string                   `json:"source_host"`
	Type       string                   `json:"type"`
	Logs       []map[string]interface{} `json:"logs"`
}

type LogSyncStats struct {
	SourceHost string `json:"source_host"`
	Type       string `json:"type"`
	Inserted   int    `json:"inserted"`
	Skipped    int    `json:"skipped"`
	Failed     int    `json:"failed"`
}

// SyncLogs accepts incremental logs pushed from nodes.
func SyncLogs(c *gin.Context) {
	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "数据库未初始化")
		return
	}
	if !checkSyncToken(c) {
		return
	}

	var req LogSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "请求解析失败: "+err.Error())
		return
	}

	req.Type = strings.TrimSpace(strings.ToLower(req.Type))
	req.SourceHost = strings.TrimSpace(req.SourceHost)

	if req.SourceHost == "" || req.Type == "" || len(req.Logs) == 0 {
		utils.ResponseError(c, http.StatusBadRequest, "source_host、type、logs 不能为空")
		return
	}

	stats := LogSyncStats{SourceHost: req.SourceHost, Type: req.Type}
	var err error

	switch req.Type {
	case "heralding_auth":
		err = syncHeraldingAuth(req, &stats)
	case "heralding_session":
		err = syncHeraldingSession(req, &stats)
	case "cowrie":
		err = syncCowrie(req, &stats)
	case "mysql_honeypot":
		err = syncMySQLHoneypot(req, &stats)
	case "dionaea":
		err = syncDionaea(req, &stats)
	default:
		utils.ResponseError(c, http.StatusBadRequest, "不支持的 type: "+req.Type)
		return
	}

	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.ResponseSuccess(c, stats)
}

func checkSyncToken(c *gin.Context) bool {
	token := c.GetHeader("X-SYNC-TOKEN")
	if token == "" {
		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			token = strings.TrimSpace(auth[len("bearer "):])
		}
	}

	if token == "" || token != config.SyncToken {
		utils.ResponseError(c, http.StatusUnauthorized, "同步令牌无效")
		return false
	}
	return true
}

func syncHeraldingAuth(req LogSyncRequest, stats *LogSyncStats) error {
	repo := repositories.NewMySQLHeraldingAuthLogRepo(config.MySQLDB)
	seen := make(map[string]struct{})

	for _, row := range req.Logs {
		var logItem repositories.HeraldingAuthLog
		if err := mapToStruct(row, &logItem); err != nil {
			stats.Failed++
			continue
		}

		if logItem.AuthID == "" {
			stats.Failed++
			continue
		}
		if _, ok := seen[logItem.AuthID]; ok {
			stats.Skipped++
			continue
		}
		seen[logItem.AuthID] = struct{}{}

		existing, err := repo.GetByAuthID(logItem.AuthID)
		if err != nil {
			stats.Failed++
			continue
		}
		if existing != nil {
			stats.Skipped++
			continue
		}

		if logItem.Timestamp.IsZero() {
			logItem.Timestamp = time.Now()
		}
		if logItem.CreatedAt.IsZero() {
			logItem.CreatedAt = time.Now()
		}

		if err := repo.Create(&logItem); err != nil {
			stats.Failed++
			continue
		}
		stats.Inserted++
	}
	return nil
}

func syncHeraldingSession(req LogSyncRequest, stats *LogSyncStats) error {
	repo := repositories.NewMySQLHeraldingSessionLogRepo(config.MySQLDB)
	seen := make(map[string]struct{})

	for _, row := range req.Logs {
		var logItem repositories.HeraldingSessionLog
		if err := mapToStruct(row, &logItem); err != nil {
			stats.Failed++
			continue
		}

		if logItem.SessionID == "" {
			stats.Failed++
			continue
		}
		if _, ok := seen[logItem.SessionID]; ok {
			stats.Skipped++
			continue
		}
		seen[logItem.SessionID] = struct{}{}

		var existing repositories.HeraldingSessionLog
		dbErr := config.MySQLDB.Where("session_id = ?", logItem.SessionID).First(&existing).Error
		if dbErr == nil {
			stats.Skipped++
			continue
		}

		if logItem.Timestamp.IsZero() {
			logItem.Timestamp = time.Now()
		}

		if err := repo.Create(&logItem); err != nil {
			stats.Failed++
			continue
		}
		stats.Inserted++
	}
	return nil
}

func syncCowrie(req LogSyncRequest, stats *LogSyncStats) error {
	repo := repositories.NewMySQLCowrieLogRepo(config.MySQLDB)
	seen := make(map[string]struct{})

	for _, row := range req.Logs {
		var logItem repositories.CowrieLog
		if err := mapToStruct(row, &logItem); err != nil {
			stats.Failed++
			continue
		}

		if logItem.AuthID == "" {
			stats.Failed++
			continue
		}
		if _, ok := seen[logItem.AuthID]; ok {
			stats.Skipped++
			continue
		}
		seen[logItem.AuthID] = struct{}{}

		existing, err := repo.GetByAuthID(logItem.AuthID)
		if err != nil {
			stats.Failed++
			continue
		}
		if existing != nil {
			stats.Skipped++
			continue
		}

		if logItem.EventTime.IsZero() {
			logItem.EventTime = time.Now()
		}
		if logItem.CreatedAt.IsZero() {
			logItem.CreatedAt = time.Now()
		}

		if err := repo.Create(&logItem); err != nil {
			stats.Failed++
			continue
		}
		stats.Inserted++
	}
	return nil
}

func syncMySQLHoneypot(req LogSyncRequest, stats *LogSyncStats) error {
	repo := repositories.NewMySQLMySQLHoneypotLogRepo(config.MySQLDB)
	seen := make(map[string]struct{})

	for _, row := range req.Logs {
		var logItem repositories.MySQLHoneypotLog
		if err := mapToStruct(row, &logItem); err != nil {
			stats.Failed++
			continue
		}

		key := buildKey(logItem.EventTime, logItem.SourceIP, logItem.DestinationIP, logItem.Query, logItem.ContainerID)
		if key == "" {
			key = strings.TrimSpace(logItem.EventID)
		}
		if key == "" {
			stats.Failed++
			continue
		}
		if _, ok := seen[key]; ok {
			stats.Skipped++
			continue
		}
		seen[key] = struct{}{}

		var existing repositories.MySQLHoneypotLog
		dbErr := config.MySQLDB.Where("event_time = ? AND source_ip = ? AND destination_ip = ? AND query = ? AND container_id = ?",
			logItem.EventTime, logItem.SourceIP, logItem.DestinationIP, logItem.Query, logItem.ContainerID).First(&existing).Error
		if dbErr == nil {
			stats.Skipped++
			continue
		}

		if logItem.CreatedAt.IsZero() {
			logItem.CreatedAt = time.Now()
		}

		if err := repo.Create(&logItem); err != nil {
			stats.Failed++
			continue
		}
		stats.Inserted++
	}
	return nil
}

func syncDionaea(req LogSyncRequest, stats *LogSyncStats) error {
	repo := repositories.NewMySQLDionaeaLogRepo(config.MySQLDB)
	seen := make(map[string]struct{})

	for _, row := range req.Logs {
		var logItem repositories.DionaeaLog
		if err := mapToStruct(row, &logItem); err != nil {
			stats.Failed++
			continue
		}

		key := buildKey(logItem.EventTime, logItem.SourceIP, logItem.DestinationIP, logItem.PayloadType, logItem.ContainerID)
		if key == "" {
			stats.Failed++
			continue
		}
		if _, ok := seen[key]; ok {
			stats.Skipped++
			continue
		}
		seen[key] = struct{}{}

		var existing repositories.DionaeaLog
		dbErr := config.MySQLDB.Where("event_time = ? AND source_ip = ? AND destination_ip = ? AND payload_type = ? AND container_id = ?",
			logItem.EventTime, logItem.SourceIP, logItem.DestinationIP, logItem.PayloadType, logItem.ContainerID).First(&existing).Error
		if dbErr == nil {
			stats.Skipped++
			continue
		}

		if logItem.CreatedAt.IsZero() {
			logItem.CreatedAt = time.Now()
		}

		if err := repo.Create(&logItem); err != nil {
			stats.Failed++
			continue
		}
		stats.Inserted++
	}
	return nil
}

func mapToStruct(data map[string]interface{}, out interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

func buildKey(ts time.Time, parts ...string) string {
	var b strings.Builder
	if !ts.IsZero() {
		b.WriteString(ts.UTC().Format(time.RFC3339Nano))
	}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		b.WriteString("|")
		b.WriteString(p)
	}
	return b.String()
}
