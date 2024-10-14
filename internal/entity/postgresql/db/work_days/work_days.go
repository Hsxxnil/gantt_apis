package work_days

import (
	"hta/internal/entity/postgresql/db/users"
	"hta/internal/interactor/models/special"
)

// Table struct is work_days database table struct
type Table struct {
	// 表ID
	ID string `gorm:"<-:create;column:id;type:uuid;not null;primaryKey;" json:"id"`
	// 工作日(陣列的字串型態)
	WorkWeek string `gorm:"column:work_week;type:text;" json:"work_week"`
	// 工作時間(陣列的字串型態)
	WorkingTime string `gorm:"column:working_time;type:text;" json:"working_time"`
	// create_users data
	CreatedByUsers users.Table `gorm:"foreignKey:ID;references:CreatedBy" json:"created_by_users,omitempty"`
	// update_users data
	UpdatedByUsers users.Table `gorm:"foreignKey:ID;references:UpdatedBy" json:"updated_by_users,omitempty"`
	// 引入後端專用
	special.Table
}

// Base struct is corresponding to work_days table structure file
type Base struct {
	// 表ID
	ID *string `json:"id,omitempty"`
	// 工作日(陣列的字串型態)
	WorkWeek *string `json:"work_week,omitempty"`
	// 工作時間(陣列的字串型態)
	WorkingTime *string `json:"working_time,omitempty"`
	// create_users data
	CreatedByUsers users.Base `json:"created_by_users,omitempty"`
	// update_users data
	UpdatedByUsers users.Base `json:"updated_by_users,omitempty"`
	// 引入後端專用
	special.Base
}

func (t *Table) TableName() string {
	return "work_days"
}
