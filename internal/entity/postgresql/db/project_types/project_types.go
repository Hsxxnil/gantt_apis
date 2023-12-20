package project_types

import (
	"hta/internal/entity/postgresql/db/users"
	"hta/internal/interactor/models/special"
)

// Table struct is project_types database table struct
type Table struct {
	// 表ID
	ID string `gorm:"<-:create;column:id;type:uuid;not null;primaryKey;" json:"id"`
	// 名稱
	Name string `gorm:"column:name;type:varchar;" json:"name"`
	// create_users data
	CreatedByUsers users.Table `gorm:"foreignKey:ID;references:CreatedBy" json:"created_by_users,omitempty"`
	// update_users data
	UpdatedByUsers users.Table `gorm:"foreignKey:ID;references:UpdatedBy" json:"updated_by_users,omitempty"`
	// 引入後端專用
	special.Table
}

// Base struct is corresponding to project_types table structure file
type Base struct {
	// 表ID
	ID *string `json:"id,omitempty"`
	// 名稱
	Name *string `json:"name,omitempty"`
	// 名稱s (後端查詢用）
	Names []*string `json:"names,omitempty"`
	// create_users data
	CreatedByUsers users.Base `json:"created_by_users,omitempty"`
	// update_users data
	UpdatedByUsers users.Base `json:"updated_by_users,omitempty"`
	// 引入後端專用
	special.Base
}

func (t *Table) TableName() string {
	return "project_types"
}
