package affiliations

import (
	"gantt/internal/entity/postgresql/db/users"
	"gantt/internal/interactor/models/special"
)

// Table struct is affiliations database table struct
type Table struct {
	// 表ID
	ID string `gorm:"<-:create;column:id;type:uuid;not null;primaryKey;" json:"id"`
	// 使用者ID
	UserID string `gorm:"column:user_id;type:uuid;not null;" json:"user_id"`
	// users data
	Users users.Table `gorm:"foreignKey:ID;references:UserID" json:"users,omitempty"`
	// 部門ID
	DeptID string `gorm:"column:dept_id;type:uuid;not null;" json:"dept_id"`
	// 職稱
	JobTitle string `gorm:"column:job_title;type:text;" json:"job_title,omitempty"`
	// 是否為主管
	IsSupervisor bool `gorm:"column:is_supervisor;type:boolean;default:false;" json:"is_supervisor,omitempty"`
	// 引入後端專用
	special.Table
}

// Base struct is corresponding to affiliations table structure file
type Base struct {
	// 表ID
	ID *string `json:"id,omitempty"`
	// 使用者ID
	UserID *string `json:"user_id,omitempty"`
	// 使用者IDs (後端查詢用)
	UserIDs []*string `json:"user_ids,omitempty"`
	// users data
	Users users.Base `json:"users,omitempty"`
	// 部門ID
	DeptID *string `json:"dept_id,omitempty"`
	// 職稱
	JobTitle *string `json:"job_title,omitempty"`
	// 是否為主管
	IsSupervisor *bool `json:"is_supervisor,omitempty"`
	// 引入後端專用
	special.Base
}

func (t *Table) TableName() string {
	return "affiliations"
}
