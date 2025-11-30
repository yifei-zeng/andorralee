package handlers

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"andorralee/pkg/utils"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	tableHoneypotInstances = "honeypot_instances"
	tableHoneypotLogs      = "honeypot_logs"
)

type syncRow struct {
	RemoteID  uint64                 `json:"remote_id"`
	CreatedAt *time.Time             `json:"created_at"`
	Data      map[string]interface{} `json:"data"`
}

type syncRequest struct {
	SourceHost string    `json:"source_host"`
	Table      string    `json:"table"`
	Rows       []syncRow `json:"rows"`
}

type syncStats struct {
	SourceHost string `json:"source_host"`
	Table      string `json:"table"`
	Inserted   int    `json:"inserted"`
	Updated    int    `json:"updated"`
	Failed     int    `json:"failed"`
}

// SyncEvents 接收远端数据增量
func SyncEvents(c *gin.Context) {
	if config.MySQLDB == nil {
		utils.ResponseError(c, http.StatusInternalServerError, "数据库未初始化")
		return
	}

	token := c.GetHeader("X-SYNC-TOKEN")
	if token == "" || token != config.SyncToken {
		utils.ResponseError(c, http.StatusUnauthorized, "同步令牌无效")
		return
	}

	var req syncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "请求解析失败: "+err.Error())
		return
	}

	req.Table = strings.ToLower(strings.TrimSpace(req.Table))
	if req.SourceHost == "" || req.Table == "" || len(req.Rows) == 0 {
		utils.ResponseError(c, http.StatusBadRequest, "source_host、table、rows 不能为空")
		return
	}

	stats := syncStats{SourceHost: req.SourceHost, Table: req.Table}

	switch req.Table {
	case tableHoneypotInstances:
		for _, row := range req.Rows {
			created, err := upsertHoneypotInstance(config.MySQLDB, req.SourceHost, row)
			if err != nil {
				stats.Failed++
				recordSyncError(req.SourceHost, req.Table, row, err)
				continue
			}
			if created {
				stats.Inserted++
			} else {
				stats.Updated++
			}
		}
	case tableHoneypotLogs:
		for _, row := range req.Rows {
			created, err := upsertHoneypotLog(config.MySQLDB, req.SourceHost, row)
			if err != nil {
				stats.Failed++
				recordSyncError(req.SourceHost, req.Table, row, err)
				continue
			}
			if created {
				stats.Inserted++
			} else {
				stats.Updated++
			}
		}
	default:
		utils.ResponseError(c, http.StatusBadRequest, "不支持的同步表: "+req.Table)
		return
	}

	utils.ResponseSuccess(c, stats)
}

func upsertHoneypotInstance(db *gorm.DB, sourceHost string, row syncRow) (bool, error) {
	if row.RemoteID == 0 {
		return false, fmt.Errorf("remote_id 不能为空")
	}
	if row.Data == nil {
		return false, fmt.Errorf("data 不能为空")
	}

	var incoming repositories.HoneypotInstance
	if err := decodeRow(row.Data, &incoming); err != nil {
		return false, err
	}

	remoteID := uint(row.RemoteID)
	incoming.ID = 0
	incoming.SourceHost = sourceHost
	incoming.RemoteID = &remoteID
	if incoming.CreateTime.IsZero() && row.CreatedAt != nil {
		incoming.CreateTime = *row.CreatedAt
	}
	if incoming.UpdateTime.IsZero() {
		incoming.UpdateTime = time.Now()
	}

	var existing repositories.HoneypotInstance
	err := db.Where("source_host = ? AND remote_id = ?", sourceHost, remoteID).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, db.Create(&incoming).Error
		}
		return false, err
	}

	if incoming.CreateTime.IsZero() {
		incoming.CreateTime = existing.CreateTime
	}
	incoming.ID = existing.ID

	return false, db.Session(&gorm.Session{FullSaveAssociations: false}).Save(&incoming).Error
}

func upsertHoneypotLog(db *gorm.DB, sourceHost string, row syncRow) (bool, error) {
	if row.RemoteID == 0 {
		return false, fmt.Errorf("remote_id 不能为空")
	}
	if row.Data == nil {
		return false, fmt.Errorf("data 不能为空")
	}

	var incoming repositories.HoneypotLog
	if err := decodeRow(row.Data, &incoming); err != nil {
		return false, err
	}

	remoteID := uint(row.RemoteID)
	incoming.ID = 0
	incoming.SourceHost = sourceHost
	incoming.RemoteID = &remoteID
	incoming.Instance = repositories.HoneypotInstance{}

	var existing repositories.HoneypotLog
	err := db.Where("source_host = ? AND remote_id = ?", sourceHost, remoteID).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, db.Omit("Instance").Create(&incoming).Error
		}
		return false, err
	}

	// 日志一般不更新，重复上报视为已存在
	return false, nil
}

func recordSyncError(sourceHost, table string, row syncRow, err error) {
	if config.MySQLDB == nil {
		return
	}

	payload, marshalErr := json.Marshal(row)
	if marshalErr != nil {
		payload = []byte(fmt.Sprintf("{\"marshal_error\":%q}", marshalErr.Error()))
	}

	rec := repositories.SyncError{
		SourceHost:   sourceHost,
		TargetTable:  table,
		Payload:      string(payload),
		ErrorMessage: err.Error(),
	}

	if dbErr := config.MySQLDB.Create(&rec).Error; dbErr != nil {
		fmt.Printf("记录同步错误失败: %v\n", dbErr)
	}
}

func decodeRow(data map[string]interface{}, out interface{}) error {
	blob, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(blob, out)
}
