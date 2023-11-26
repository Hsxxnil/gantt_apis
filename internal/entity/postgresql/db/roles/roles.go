package roles

import (
	"hta/internal/interactor/models/special"
)

// Table struct is roles database table struct
type Table struct {
	// 表ID
	ID string `gorm:"<-:create;column:id;type:uuid;not null;primaryKey;" json:"id"`
	// 角色名稱
	Name string `gorm:"column:name;type:text;not null;" json:"name"`
	// 角色顯示名稱
	DisplayName string `gorm:"column:display_name;type:text;not null;" json:"display_name"`
	// 角色是否啟用
	IsEnable bool `gorm:"column:is_enable;type:bool;not null;" json:"is_enable"`
	//// 引入後端專用
	special.Table
}

// Base struct is corresponding to roles table structure file
type Base struct {
	// 表ID
	ID *string `json:"id,omitempty"`
	// 角色名稱
	Name *string `json:"name,omitempty"`
	// 角色顯示名稱
	DisplayName *string `json:"display_name,omitempty"`
	// 角色是否啟用
	IsEnable *bool `json:"is_enable,omitempty"`
	//// 引入後端專用
	special.Base
}

// TableName sets the insert table name for this struct type
func (t *Table) TableName() string {
	return "roles"
}
