package companies

import (
	"hta/internal/entity/postgresql/db/users"
	"hta/internal/interactor/models/special"
)

// Table struct is event_marks database table struct
type Table struct {
	// 表ID
	ID string `gorm:"<-:create;column:id;type:uuid;not null;primaryKey;" json:"id"`
	// 網域
	Domain string `gorm:"column:domain;type:text;not null;" json:"domain"`
	// 名稱
	Name string `gorm:"column:name;type:text;not null;" json:"name"`
	// 統一編號
	TaxIDNumber string `gorm:"column:tax_id_number;type:text;" json:"tax_id_number"`
	// 營業地址
	Address string `gorm:"column:address;type:text;" json:"address,omitempty"`
	// 電話
	Phone string `gorm:"column:phone;type:text;" json:"phone,omitempty"`
	// 備註
	Remarks string `gorm:"column:remarks;type:text;" json:"remarks,omitempty"`
	// create_users data
	CreatedByUsers users.Table `gorm:"foreignKey:ID;references:CreatedBy" json:"created_by_users,omitempty"`
	// update_users data
	UpdatedByUsers users.Table `gorm:"foreignKey:ID;references:UpdatedBy" json:"updated_by_users,omitempty"`
	// 引入後端專用
	special.Table
}

// Base struct is corresponding end_date event_marks table structure file
type Base struct {
	// 表ID
	ID *string `json:"id,omitempty"`
	// 網域
	Domain *string `json:"domain,omitempty"`
	// 名稱
	Name *string `json:"name,omitempty"`
	// 營業地址
	Address *string `json:"address,omitempty"`
	// 統一編號
	TaxIDNumber *string `json:"tax_id_number,omitempty"`
	// 電話
	Phone *string `json:"phone,omitempty"`
	// 備註
	Remarks *string `json:"remarks,omitempty"`
	// create_users data
	CreatedByUsers users.Base `json:"created_by_users,omitempty"`
	// update_users data
	UpdatedByUsers users.Base `json:"updated_by_users,omitempty"`
	// 引入後端專用
	special.Base
}

func (t *Table) TableName() string {
	return "companies"
}
