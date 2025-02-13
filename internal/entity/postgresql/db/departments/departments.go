package departments

import (
	"gantt/internal/entity/postgresql/db/affiliations"
	"gantt/internal/entity/postgresql/db/users"
	"gantt/internal/interactor/models/special"
)

// Table struct is departments database table struct
type Table struct {
	// 表ID
	ID string `gorm:"<-:create;column:id;type:uuid;not null;primaryKey;" json:"id"`
	// 名稱
	Name string `gorm:"column:name;type:text;not null;" json:"name"`
	// 傳真
	Fax string `gorm:"column:fax;type:text;" json:"fax"`
	// 電話
	Tel string `gorm:"column:tel;type:text;" json:"tel,omitempty"`
	// affiliations data
	Affiliations []affiliations.Table `gorm:"foreignKey:DeptID;references:ID;" json:"affiliations,omitempty"`
	// create_users data
	CreatedByUsers users.Table `gorm:"foreignKey:ID;references:CreatedBy" json:"created_by_users,omitempty"`
	// update_users data
	UpdatedByUsers users.Table `gorm:"foreignKey:ID;references:UpdatedBy" json:"updated_by_users,omitempty"`
	// 引入後端專用
	special.Table
}

// Base struct is corresponding to departments table structure file
type Base struct {
	// 表ID
	ID *string `json:"id,omitempty"`
	// 部門IDs (後端查詢用)
	DeptIDs []*string `json:"dept_ids,omitempty"`
	// 名稱
	Name *string `json:"name,omitempty"`
	// 傳真
	Fax *string `json:"fax,omitempty"`
	// 電話
	Tel *string `json:"tel,omitempty"`
	// affiliations data
	Affiliations []affiliations.Base `json:"affiliations,omitempty"`
	// create_users data
	CreatedByUsers users.Base `json:"created_by_users,omitempty"`
	// update_users data
	UpdatedByUsers users.Base `json:"updated_by_users,omitempty"`
	// 引入後端專用
	special.Base
}

func (t *Table) TableName() string {
	return "departments"
}
