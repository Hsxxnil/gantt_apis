package task_resources

import (
	"hta/internal/entity/postgresql/db/project_resources"
	"hta/internal/entity/postgresql/db/users"
	"hta/internal/interactor/models/special"
)

// Table struct is task_resources database table struct
type Table struct {
	// 表ID
	ID string `gorm:"<-:create;column:id;type:uuid;not null;primaryKey;" json:"id"`
	// 任務UUID
	TaskUUID string `gorm:"column:task_uuid;type:uuid;not null;" json:"task_uuid"`
	// 資源UUID
	ResourceUUID string `gorm:"column:resource_uuid;type:uuid;not null;" json:"resource_uuid"`
	// project_resources data
	Resources project_resources.Table `gorm:"foreignKey:ResourceUUID;references:ResourceUUID" json:"resources,omitempty"`
	// 單位
	Unit float64 `gorm:"column:unit;type:numeric;not null;" json:"unit"`
	// create_users data
	CreatedByUsers users.Table `gorm:"foreignKey:ID;references:CreatedBy" json:"created_by_users,omitempty"`
	// update_users data
	UpdatedByUsers users.Table `gorm:"foreignKey:ID;references:UpdatedBy" json:"updated_by_users,omitempty"`
	// 引入後端專用
	special.Table
}

// Base struct is corresponding to task_resources table structure file
type Base struct {
	// 表ID
	ID *string `json:"id,omitempty"`
	// 任務UUID
	TaskUUID *string `json:"task_uuid,omitempty"`
	// 資源UUID
	ResourceUUID *string `json:"resource_uuid,omitempty"`
	// 單位
	Unit *float64 `json:"unit,omitempty"`
	// 任務UUIDs (後端批量刪除用）
	TaskUUIDs []*string `json:"task_uuids,omitempty"`
	// 資源UUIDs (後端批量刪除用）
	ResourceUUIDs []*string `json:"resource_uuids,omitempty"`
	// project_resources data
	Resources project_resources.Base `json:"resources,omitempty"`
	// create_users data
	CreatedByUsers users.Base `json:"created_by_users,omitempty"`
	// update_users data
	UpdatedByUsers users.Base `json:"updated_by_users,omitempty"`
	// 引入後端專用
	special.Base
}

func (t *Table) TableName() string {
	return "task_resources"
}
