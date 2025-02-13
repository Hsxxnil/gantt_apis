package holidays

import (
	"gantt/internal/entity/postgresql/db/users"
	"gantt/internal/interactor/models/special"
	"time"
)

// Table struct is holidays database table struct
type Table struct {
	// 表ID
	ID string `gorm:"<-:create;column:id;type:uuid;not null;primaryKey;" json:"id"`
	// 名稱
	Name string `gorm:"column:name;type:text;" json:"label"`
	// 起始日期
	StartDate *time.Time `gorm:"column:start_date;type:timestamp;" json:"from"`
	// 結束日期
	EndDate *time.Time `gorm:"column:end_date;type:timestamp;" json:"to"`
	// 前端css
	Css string `gorm:"column:css;type:text;" json:"cssClass"`
	// create_users data
	CreatedByUsers users.Table `gorm:"foreignKey:ID;references:CreatedBy" json:"created_by_users,omitempty"`
	// update_users data
	UpdatedByUsers users.Table `gorm:"foreignKey:ID;references:UpdatedBy" json:"updated_by_users,omitempty"`
	// 引入後端專用
	special.Table
}

// Base struct is corresponding to holidays table structure file
type Base struct {
	// 表ID
	ID *string `json:"id,omitempty"`
	// 名稱
	Name *string `json:"label,omitempty"`
	// 起始日期
	StartDate *time.Time `json:"from,omitempty"`
	// 結束日期
	EndDate *time.Time `json:"to,omitempty"`
	// 前端css
	Css *string `json:"cssClass,omitempty"`
	// create_users data
	CreatedByUsers users.Base `json:"created_by_users,omitempty"`
	// update_users data
	UpdatedByUsers users.Base `json:"updated_by_users,omitempty"`
	// 引入後端專用
	special.Base
}

func (t *Table) TableName() string {
	return "holidays"
}
