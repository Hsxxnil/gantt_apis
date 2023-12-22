package departments

import (
	"hta/internal/entity/postgresql/db/users"
	"hta/internal/interactor/models/special"
)

// Table struct is departments database table struct
type Table struct {
	// 表ID
	ID string `gorm:"<-:create;column:id;type:uuid;not null;primaryKey;" json:"id"`
	// 部門主管ID(user_id)
	SupervisorID *string `gorm:"column:supervisor_id;type:uuid;" json:"supervisor_id"`
	// 名稱
	Name string `gorm:"column:name;type:text;not null;" json:"name"`
	// 傳真
	Fax string `gorm:"column:fax;type:text;" json:"fax"`
	// 電話
	Tel string `gorm:"column:tel;type:text;" json:"tel,omitempty"`
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
	// 部門主管ID(user_id)
	SupervisorID *string `json:"supervisor_id,omitempty"`
	// 名稱
	Name *string `json:"name,omitempty"`
	// 傳真
	Fax *string `json:"fax,omitempty"`
	// 電話
	Tel *string `json:"tel,omitempty"`
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
