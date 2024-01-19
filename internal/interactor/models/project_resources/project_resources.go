package project_resources

import (
	"hta/internal/interactor/models/page"
	"hta/internal/interactor/models/section"
)

// Create struct is used to create achieves
type Create struct {
	// 專案UUID
	ProjectUUID string `json:"project_uuid,omitempty" binding:"required,uuid4" validate:"required,uuid4"`
	// 資源UUID
	ResourceUUID string `json:"resource_uuid,omitempty" binding:"required,uuid4" validate:"required,uuid4"`
	// 專案角色
	Role string `json:"role,omitempty"`
	// 是否可編輯專案任務
	IsEditable bool `json:"is_editable"`
	// 創建者
	CreatedBy string `json:"created_by,omitempty" binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
}

// Field is structure file for search
type Field struct {
	// 表ID
	ID string `json:"id,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4" swaggerignore:"true"`
	// 專案UUID
	ProjectUUID *string `json:"project_uuid,omitempty" form:"project_uuid"`
	// 專案UUIDs (後端查詢用）
	ProjectUUIDs []*string `json:"project_uuids,omitempty" form:"project_uuids" swaggerignore:"true"`
	// 資源UUID
	ResourceUUID *string `json:"resource_uuid,omitempty" form:"resource_uuid"`
	// 資源UUIDs (後端查詢用）
	ResourceUUIDs []*string `json:"resource_uuids,omitempty" form:"resource_uuids" swaggerignore:"true"`
	// 專案角色
	Role *string `json:"role,omitempty" form:"role"`
	// 是否可編輯專案任務
	IsEditable *bool `json:"is_editable" form:"is_editable"`
	// 搜尋欄位
	Filter `json:"filter"`
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
	// 角色
	FilterRole string `json:"role,omitempty"`
	// 群組
	FilterResourceGroup string `json:"resource_group,omitempty"`
}

// List is multiple return structure files
type List struct {
	// 多筆
	ProjectResources []Single `json:"project_resources"`
	// 分頁返回結構檔
	page.Total
}

// Single return structure file
type Single struct {
	// 表ID
	ID string `json:"id,omitempty"`
	// 專案UUID
	ProjectUUID string `json:"project_uuid,omitempty"`
	// 資源UUID
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
	ResourceGroup string `json:"resource_group,omitempty"`
	//
	IsExpand bool `json:"is_expand,omitempty"`
	// 專案角色
	Role string `json:"role,omitempty"`
	// 是否可編輯專案任務
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
	ID string `json:"id,omitempty"  binding:"omitempty,uuid4" validate:"omitempty,uuid4" swaggerignore:"true"`
	// 專案角色
	Role *string `json:"role,omitempty"`
	// 是否可編輯專案任務
	IsEditable *bool `json:"is_editable"`
	// 更新者
	UpdatedBy *string `json:"updated_by,omitempty" binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
}

// ProjectIDs struct is used to get multiple project data
type ProjectIDs struct {
	// 多筆
	Projects []*string `json:"projects"`
	// 搜尋欄位
	Filter `json:"filter"`
}
