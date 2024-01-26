package resources

import (
	"encoding/csv"
	"hta/internal/interactor/models/page"
	"hta/internal/interactor/models/section"
	"hta/internal/interactor/models/sort"
)

// Create struct is used to create achieves
type Create struct {
	// 名字
	ResourceName string `json:"resource_name,omitempty"`
	// 信箱
	Email string `json:"email,omitempty" binding:"required,email" validate:"required,email"`
	// 電話
	Phone string `json:"phone,omitempty"`
	//
	StandardCost float64 `json:"standard_cost,omitempty" binding:"omitempty,gte=0" validate:"omitempty,gte=0"`
	//
	TotalCost float64 `json:"total_cost,omitempty" binding:"omitempty,gte=0" validate:"omitempty,gte=0"`
	// 總負載
	TotalLoad float64 `json:"total_load,omitempty" binding:"omitempty,gte=0" validate:"omitempty,gte=0"`
	// 群組(後端寫入)
	ResourceGroup string `json:"resource_group,omitempty" swaggerignore:"true"`
	// 群組
	ResourceGroups []string `json:"resource_groups,omitempty"`
	//
	IsExpand bool `json:"is_expand,omitempty"`
	// 創建者
	CreatedBy string `json:"created_by,omitempty" binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
}

// Field is structure file for search
type Field struct {
	// 表ID
	ResourceUUID string `json:"resource_uuid,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4" swaggerignore:"true"`
	// 表IDs (後端查詢用)
	ResourceUUIDs []*string `json:"resource_uuids,omitempty" form:"resource_uuids" swaggerignore:"true"`
	// 舊編號 ＆ 頁面ID
	ResourceID *int64 `json:"resource_id,omitempty" form:"resource_id"`
	// 名字
	ResourceName *string `json:"resource_name,omitempty" form:"resource_name"`
	// 信箱
	Email *string `json:"email,omitempty" form:"email"`
	// 電話
	Phone *string `json:"phone,omitempty" form:"phone"`
	//
	StandardCost *float64 `json:"standard_cost,omitempty" form:"standard_cost"`
	//
	TotalCost *float64 `json:"total_cost,omitempty" form:"total_cost"`
	// 總負載
	TotalLoad *float64 `json:"total_load,omitempty" form:"total_load"`
	// 群組
	ResourceGroup *string `json:"resource_group,omitempty" form:"resource_group"`
	//
	IsExpand *bool `json:"is_expand,omitempty" form:"is_expand"`
	// 創建者
	CreatedBy *string `json:"created_by,omitempty" form:"created_by"`
	// 使用者角色
	Role *string `json:"role,omitempty" form:"role"`
	// 搜尋欄位
	Filter `json:"filter"`
	// 排序欄位
	sort.Sort `json:"sort"`
}

// Fields is the searched structure file (including pagination)
type Fields struct {
	// 搜尋結構檔
	Field
	// 分頁搜尋結構檔
	page.Pagination
}

// Filter struct is used to store the search field
type Filter struct {
	// 名字
	FilterResourceName string `json:"resource_name,omitempty"`
	// 群組
	FilterResourceGroups []string `json:"resource_groups,omitempty"`
	// 信箱
	FilterEmail string `json:"email,omitempty"`
	// 電話
	FilterPhone string `json:"phone,omitempty"`
}

// List is multiple return structure files
type List struct {
	// 多筆
	Resources []*struct {
		// 表ID
		ResourceUUID string `json:"resource_uuid,omitempty"`
		// 舊編號 ＆ 頁面ID
		ResourceID int64 `json:"resource_id"`
		// 名字
		ResourceName string `json:"resource_name,omitempty"`
		// 信箱
		Email string `json:"email,omitempty"`
		// 電話
		Phone string `json:"phone,omitempty"`
		//
		StandardCost float64 `json:"standard_cost,omitempty"`
		//
		TotalCost float64 `json:"total_cost,omitempty"`
		// 總負載
		TotalLoad float64 `json:"total_load,omitempty"`
		// 群組
		ResourceGroups []string `json:"resource_groups,omitempty"`
		//
		IsExpand bool `json:"is_expand,omitempty"`
		// 是否綁定
		IsBind bool `json:"is_bind"`
		// 是否可編輯或刪除資源
		IsEditable bool `json:"is_editable"`
		// 創建者
		CreatedBy string `json:"created_by,omitempty"`
		// 更新者
		UpdatedBy string `json:"updated_by,omitempty"`
		// 時間戳記
		section.TimeAt
	} `json:"resources"`
	// 分頁返回結構檔
	page.Total
}

// Single return structure file
type Single struct {
	// 表ID
	ResourceUUID string `json:"resource_uuid,omitempty"`
	// 舊編號 ＆ 頁面ID
	ResourceID int64 `json:"resource_id"`
	// 名字
	ResourceName string `json:"resource_name,omitempty"`
	// 信箱
	Email string `json:"email,omitempty"`
	// 電話
	Phone string `json:"phone,omitempty"`
	//
	StandardCost float64 `json:"standard_cost,omitempty"`
	//
	TotalCost float64 `json:"total_cost,omitempty"`
	// 總負載
	TotalLoad float64 `json:"total_load,omitempty"`
	// 群組
	ResourceGroups []string `json:"resource_groups,omitempty"`
	//
	IsExpand bool `json:"is_expand,omitempty"`
	// 是否綁定
	IsBind bool `json:"is_bind"`
	// 是否可編輯或刪除資源
	IsEditable bool `json:"is_editable"`
	// 創建者
	CreatedBy string `json:"created_by,omitempty"`
	// 更新者
	UpdatedBy string `json:"updated_by,omitempty"`
	// 時間戳記
	section.TimeAt
}

// Update struct is used to update achieves
type Update struct {
	// 表ID
	ResourceUUID string `json:"resource_uuid,omitempty"  binding:"omitempty,uuid4" validate:"omitempty,uuid4" swaggerignore:"true"`
	// 名字
	ResourceName *string `json:"resource_name,omitempty"`
	// 信箱
	Email *string `json:"email,omitempty"`
	// 電話
	Phone *string `json:"phone,omitempty"`
	//
	StandardCost *float64 `json:"standard_cost,omitempty" binding:"omitempty,gte=0" validate:"omitempty,gte=0"`
	//
	TotalCost *float64 `json:"total_cost,omitempty" binding:"omitempty,gte=0" validate:"omitempty,gte=0"`
	// 總負載
	TotalLoad *float64 `json:"total_load,omitempty" binding:"omitempty,gte=0" validate:"omitempty,gte=0"`
	// 群組(後端寫入)
	ResourceGroup *string `json:"resource_group,omitempty" swaggerignore:"true"`
	// 群組
	ResourceGroups []*string `json:"resource_groups,omitempty"`
	//
	IsExpand *bool `json:"is_expand,omitempty"`
	// 更新者
	UpdatedBy *string `json:"updated_by,omitempty" binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
	// 使用者角色
	Role *string `json:"role,omitempty" swaggerignore:"true"`
}

// TaskSingle return structure file for tasks
type TaskSingle struct {
	// 表ID
	ResourceUUID string `json:"resource_uuid,omitempty"`
	// 舊編號 ＆ 頁面ID
	ResourceID int64 `json:"resource_id"`
	// 名字
	ResourceName string `json:"resource_name,omitempty"`
	// 信箱
	Email string `json:"email,omitempty"`
	// 電話
	Phone string `json:"phone,omitempty"`
	//
	StandardCost float64 `json:"standard_cost,omitempty"`
	//
	TotalCost float64 `json:"total_cost,omitempty"`
	// 總負載
	TotalLoad float64 `json:"total_load,omitempty"`
	// 群組
	ResourceGroups []string `json:"resource_groups,omitempty"`
	//
	IsExpand bool `json:"is_expand,omitempty"`
	// 單位
	Unit float64 `json:"unit,omitempty"`
	// 專案角色
	Role string `json:"role,omitempty"`
}

// Import struct is used to import the task file
type Import struct {
	// CSV檔案
	CSVFile *csv.Reader `swaggerignore:"true"`
	// Base64
	Base64 string `json:"base64,omitempty" binding:"required,base64" validate:"required,base64"`
	// 創建者
	CreatedBy string `json:"created_by,omitempty" binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
}
