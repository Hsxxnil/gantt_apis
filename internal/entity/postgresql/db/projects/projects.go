package projects

import (
	"gantt/internal/entity/postgresql/db/project_types"
	"gantt/internal/entity/postgresql/db/users"
	model "gantt/internal/interactor/models/projects"
	"gantt/internal/interactor/models/special"
	"time"
)

// Table struct is projects database table struct
type Table struct {
	// 表ID
	ProjectUUID string `gorm:"<-:create;column:project_uuid;type:uuid;not null;primaryKey;" json:"project_uuid"`
	// 前端編號 (非表ID)
	ProjectID string `gorm:"->;column:project_id;type:text;" json:"project_id"`
	// 名稱
	ProjectName string `gorm:"column:project_name;type:text;" json:"project_name"`
	// 類別
	TypeID string `gorm:"column:type_id;type:uuid;" json:"type_id"`
	// project_types data
	ProjectTypes project_types.Table `gorm:"foreignKey:ID;references:TypeID" json:"project_types,omitempty"`
	// 代號
	Code string `gorm:"column:code;type:text;" json:"code"`
	// 起始日期
	StartDate *time.Time `gorm:"column:start_date;type:timestamp;" json:"start_date"`
	// 結束日期
	EndDate *time.Time `gorm:"column:end_date;type:timestamp;" json:"end_date"`
	// 客戶
	Client string `gorm:"column:client;type:text;" json:"client"`
	// 狀態
	Status string `gorm:"column:status;type:text;" json:"status"`
	// create_users data
	CreatedByUsers users.Table `gorm:"foreignKey:ID;references:CreatedBy" json:"created_by_users,omitempty"`
	// update_users data
	UpdatedByUsers users.Table `gorm:"foreignKey:ID;references:UpdatedBy" json:"updated_by_users,omitempty"`
	// 引入後端專用
	special.Table
}

// Base struct is corresponding to projects table structure file
type Base struct {
	// 表ID
	ProjectUUID *string `json:"project_uuid,omitempty"`
	// 專案UUIDs (後端查詢用）
	ProjectUUIDs []string `json:"project_uuids,omitempty"`
	// 前端編號 (非表ID)
	ProjectID *string `json:"project_id,omitempty"`
	// 名稱
	ProjectName *string `json:"project_name,omitempty"`
	// 類別
	TypeID *string `json:"type_id,omitempty"`
	// project_types data
	ProjectTypes project_types.Base `json:"project_types,omitempty"`
	// 代號
	Code *string `json:"code,omitempty"`
	// 起始日期
	StartDate *time.Time `json:"start_date,omitempty"`
	// 結束日期
	EndDate *time.Time `json:"end_date,omitempty"`
	// 客戶
	Client *string `json:"client,omitempty"`
	// 狀態
	Status *string `json:"status,omitempty"`
	// create_users data
	CreatedByUsers users.Base `json:"created_by_users,omitempty"`
	// update_users data
	UpdatedByUsers users.Base `json:"updated_by_users,omitempty"`
	// 引入後端專用
	special.Base
	// 搜尋欄位
	model.Filter `json:"filter"`
}

func (t *Table) TableName() string {
	return "projects"
}
