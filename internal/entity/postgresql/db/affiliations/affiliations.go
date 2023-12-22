package affiliations

import (
	"hta/internal/entity/postgresql/db/users"
	"hta/internal/interactor/models/special"
)

// Table struct is affiliations database table struct
type Table struct {
	// 表ID
	ID string `gorm:"<-:create;column:id;type:uuid;not null;primaryKey;" json:"id"`
	// 使用者ID
	UserID string `gorm:"column:user_id;type:uuid;not null;" json:"user_id"`
	// 部門ID
	DeptID string `gorm:"column:dept_id;type:uuid;not null;" json:"dept_id"`
	// 職稱
	JobTitle string `gorm:"column:job_title;type:text;" json:"job_title,omitempty"`
	// 是否為主管
	IsSupervisor bool `gorm:"column:is_supervisor;type:boolean;default:false;" json:"is_supervisor,omitempty"`
	// create_users data
	CreatedByUsers users.Table `gorm:"foreignKey:ID;references:CreatedBy" json:"created_by_users,omitempty"`
	// update_users data
	UpdatedByUsers users.Table `gorm:"foreignKey:ID;references:UpdatedBy" json:"updated_by_users,omitempty"`
	// 引入後端專用
	special.Table
}

// Base struct is corresponding to affiliations table structure file
type Base struct {
	// 表ID
	ID *string `json:"id,omitempty"`
	// 使用者ID
	UserID *string `json:"user_id,omitempty"`
	// 部門ID
	DeptID *string `json:"dept_id,omitempty"`
	// 職稱
	JobTitle *string `json:"job_title,omitempty"`
	// 是否為主管
	IsSupervisor *bool `json:"is_supervisor,omitempty"`
	// create_users data
	CreatedByUsers users.Base `json:"created_by_users,omitempty"`
	// update_users data
	UpdatedByUsers users.Base `json:"updated_by_users,omitempty"`
	// 引入後端專用
	special.Base
}

func (t *Table) TableName() string {
	return "affiliations"
}
