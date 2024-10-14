package project_resources

import (
	"hta/internal/entity/postgresql/db/resources"
	"hta/internal/entity/postgresql/db/users"
	model "hta/internal/interactor/models/project_resources"
	"hta/internal/interactor/models/special"
)

// Table struct is project_resources database table struct
type Table struct {
	// 表ID
	ID string `gorm:"<-:create;column:id;type:uuid;not null;primaryKey;" json:"id"`
	// 專案UUID
	ProjectUUID string `gorm:"column:project_uuid;type:uuid;not null;" json:"project_uuid"`
	// 資源UUID
	ResourceUUID string `gorm:"column:resource_uuid;type:uuid;not null;" json:"resource_uuid"`
	// resources data
	Resources resources.Table `gorm:"foreignKey:ResourceUUID;references:ResourceUUID" json:"resources,omitempty"`
	// 專案角色
	Role string `gorm:"column:role;type:text;" json:"role"`
	// 是否可編輯專案任務
	IsEditable *bool `gorm:"column:is_editable;type:boolean;default:true;" json:"is_editable"`
	// create_users data
	CreatedByUsers users.Table `gorm:"foreignKey:ID;references:CreatedBy" json:"created_by_users,omitempty"`
	// update_users data
	UpdatedByUsers users.Table `gorm:"foreignKey:ID;references:UpdatedBy" json:"updated_by_users,omitempty"`
	// 引入後端專用
	special.Table
}

// Base struct is corresponding to project_resources table structure file
type Base struct {
	// 表ID
	ID *string `json:"id,omitempty"`
	// 專案UUID
	ProjectUUID *string `json:"project_uuid,omitempty"`
	// 專案UUIDs (後端查詢用）
	ProjectUUIDs []*string `json:"project_uuids,omitempty"`
	// 資源UUID
	ResourceUUID *string `json:"resource_uuid,omitempty"`
	// 資源UUIDs (後端查詢用）
	ResourceUUIDs []*string `json:"resource_uuids,omitempty"`
	// 專案角色
	Role *string `json:"role,omitempty"`
	// 是否可編輯專案任務
	IsEditable *bool `json:"is_editable"`
	// resources data
	Resources resources.Base `json:"resources,omitempty"`
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
	return "project_resources"
}
