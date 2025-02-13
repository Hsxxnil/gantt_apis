package project_types

import (
	"gantt/internal/interactor/models/special"
)

// Table struct is project_types database table struct
type Table struct {
	// 表ID
	ID string `gorm:"<-:create;column:id;type:uuid;not null;primaryKey;" json:"id"`
	// 名稱
	Name string `gorm:"column:name;type:text;" json:"name"`
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
	// 引入後端專用
	special.Base
}

func (t *Table) TableName() string {
	return "project_types"
}
