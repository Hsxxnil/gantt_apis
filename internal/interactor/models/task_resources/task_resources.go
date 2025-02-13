package task_resources

import (
	"gantt/internal/interactor/models/page"
	"gantt/internal/interactor/models/section"
)

// Create struct is used to create achieves
type Create struct {
	// 任務UUID
	TaskUUID string `json:"task_uuid,omitempty" binding:"required,uuid4" validate:"required,uuid4"`
	// 資源UUID
	ResourceUUID string `json:"resource_uuid,omitempty" binding:"required,uuid4" validate:"required,uuid4"`
	// 單位
	Unit float64 `json:"unit,omitempty" binding:"required,gte=0" validate:"required,gte=0"`
	// 創建者
	CreatedBy string `json:"created_by,omitempty" binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
}

// Field is structure file for search
type Field struct {
	// 表ID
	ID string `json:"id,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4" swaggerignore:"true"`
	// 任務UUID
	TaskUUID *string `json:"task_uuid,omitempty" form:"task_uuid"`
	// 資源UUID
	ResourceUUID *string `json:"resource_uuid,omitempty" form:"resource_uuid"`
	// 任務UUIDs (後端批量刪除用）
	TaskUUIDs []*string `json:"task_uuids,omitempty" form:"task_uuids" swaggerignore:"true"`
	// 資源UUIDs (後端批量刪除用）
	ResourceUUIDs []*string `json:"resource_uuids,omitempty" form:"resource_uuids" swaggerignore:"true"`
}

// Fields is the searched structure file (including pagination)
type Fields struct {
	// 搜尋結構檔
	Field
	// 分頁搜尋結構檔
	page.Pagination
}

// List is multiple return structure files
type List struct {
	// 多筆
	Units []*struct {
		// 表ID
		ID string `json:"id,omitempty"`
		// 任務UUID
		TaskUUID string `json:"task_uuid,omitempty"`
		// 資源UUID
		ResourceUUID string `json:"resource_uuid,omitempty"`
		// 單位
		Unit float64 `json:"unit,omitempty"`
		// 創建者
		CreatedBy string `json:"created_by,omitempty"`
		// 更新者
		UpdatedBy string `json:"updated_by,omitempty"`
		// 時間戳記
		section.TimeAt
	} `json:"task_resources"`
	// 分頁返回結構檔
	page.Total
}

// Single return structure file
type Single struct {
	// 表ID
	ID string `json:"id,omitempty"`
	// 任務UUID
	TaskUUID string `json:"task_uuid,omitempty"`
	// 資源UUID
	ResourceUUID string `json:"resource_uuid,omitempty"`
	// 單位
	Unit float64 `json:"unit,omitempty"`
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
	// 單位
	Unit *float64 `json:"unit,omitempty" binding:"omitempty,gte=0" validate:"omitempty,gte=0"`
	// 更新者
	UpdatedBy *string `json:"updated_by,omitempty" binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
}
