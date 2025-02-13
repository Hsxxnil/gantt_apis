package event_marks

import (
	"gantt/internal/entity/postgresql/db/users"
	"gantt/internal/interactor/models/special"
	"time"
)

// Table struct is event_marks database table struct
type Table struct {
	// 表ID
	ID string `gorm:"<-:create;column:id;type:uuid;not null;primaryKey;" json:"id"`
	// 名稱
	Name string `gorm:"column:name;type:text;" json:"label"`
	// 日期
	Day *time.Time `gorm:"column:day;type:timestamp;" json:"day"`
	// 專案UUID
	ProjectUUID *string `gorm:"column:project_uuid;type:uuid;" json:"project_uuid"`
	// create_users data
	CreatedByUsers users.Table `gorm:"foreignKey:ID;references:CreatedBy" json:"created_by_users,omitempty"`
	// update_users data
	UpdatedByUsers users.Table `gorm:"foreignKey:ID;references:UpdatedBy" json:"updated_by_users,omitempty"`
	// 引入後端專用
	special.Table
}

// Base struct is corresponding to event_marks table structure file
type Base struct {
	// 表ID
	ID *string `json:"id,omitempty"`
	// 名稱
	Name *string `json:"label,omitempty"`
	// 起始日期
	Day *time.Time `json:"day,omitempty"`
	// 專案UUID
	ProjectUUID *string `json:"project_uuid,omitempty"`
	// create_users data
	CreatedByUsers users.Base `json:"created_by_users,omitempty"`
	// update_users data
	UpdatedByUsers users.Base `json:"updated_by_users,omitempty"`
	// 引入後端專用
	special.Base
}

func (t *Table) TableName() string {
	return "event_marks"
}
