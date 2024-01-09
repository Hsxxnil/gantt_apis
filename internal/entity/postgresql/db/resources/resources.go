package resources

import (
	"hta/internal/entity/postgresql/db/users"
	model "hta/internal/interactor/models/resources"
	"hta/internal/interactor/models/sort"
	"hta/internal/interactor/models/special"
)

// Table struct is resources database table struct
type Table struct {
	// 表ID
	ResourceUUID string `gorm:"<-:create;column:resource_uuid;type:uuid;not null;primaryKey;" json:"resource_uuid"`
	// 舊編號 ＆ 頁面ID (非表ID)
	ResourceID int64 `gorm:"->;column:resource_id;type:serial;" json:"resource_id"`
	// 名字
	ResourceName string `gorm:"column:resource_name;type:varchar;" json:"resource_name"`
	// 信箱
	Email string `gorm:"column:email;type:varchar;" json:"email"`
	// 電話
	Phone string `gorm:"column:phone;type:varchar;" json:"phone"`
	//
	StandardCost float64 `gorm:"column:standard_cost;type:numeric" json:"standard_cost"`
	//
	TotalCost float64 `gorm:"column:total_cost;type:numeric" json:"total_cost"`
	// 總負載
	TotalLoad float64 `gorm:"column:total_load;type:numeric" json:"total_load"`
	// 群組
	ResourceGroup string `gorm:"column:resource_group;type:varchar;" json:"resource_group"`
	//
	IsExpand bool `gorm:"column:is_expand;type:boolean;default:false" json:"is_expand"`
	// create_users data
	CreatedByUsers users.Table `gorm:"foreignKey:ID;references:CreatedBy" json:"created_by_users,omitempty"`
	// update_users data
	UpdatedByUsers users.Table `gorm:"foreignKey:ID;references:UpdatedBy" json:"updated_by_users,omitempty"`
	// 引入後端專用
	special.Table
}

// Base struct is corresponding to resources table structure file
type Base struct {
	// 表ID
	ResourceUUID *string `json:"resource_uuid,omitempty"`
	// 舊編號 ＆ 頁面ID
	ResourceID *int64 `json:"resource_id,omitempty"`
	// 名字
	ResourceName *string `json:"resource_name,omitempty"`
	// 信箱
	Email *string `json:"email,omitempty"`
	// 電話
	Phone *string `json:"phone,omitempty"`
	//
	StandardCost *float64 `json:"standard_cost,omitempty"`
	//
	TotalCost *float64 `json:"total_cost,omitempty"`
	// 總負載
	TotalLoad *float64 `json:"total_load,omitempty"`
	// 群組
	ResourceGroup *string `json:"resource_group,omitempty"`
	//
	IsExpand *bool `json:"is_expand,omitempty"`
	// create_users data
	CreatedByUsers users.Base `json:"created_by_users,omitempty"`
	// update_users data
	UpdatedByUsers users.Base `json:"updated_by_users,omitempty"`
	// 引入後端專用
	special.Base
	// 排序欄位
	sort.Sort `json:"sort"`
	// 搜尋欄位
	model.Filter `json:"filter"`
}

func (t *Table) TableName() string {
	return "resources"
}
