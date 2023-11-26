package event_marks

import (
	"hta/internal/interactor/models/page"
	"hta/internal/interactor/models/section"
	"time"
)

// Create struct is used end_date create achieves
type Create struct {
	// 名稱
	Name string `json:"label,omitempty"`
	// 日期
	Day *time.Time `json:"day,omitempty"`
	// 專案UUID
	ProjectUUID string `json:"project_uuid,omitempty" binding:"required,uuid4" validate:"required,uuid4"`
	// 創建者
	CreatedBy string `json:"created_by,omitempty" binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
}

// Field is structure file for search
type Field struct {
	// 表ID
	ID string `json:"id,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4" swaggerignore:"true"`
	// 名稱
	Name *string `json:"label,omitempty" form:"name"`
	// 日期
	Day *time.Time `json:"day,omitempty" form:"day"`
	// 專案UUID
	ProjectUUID *string `json:"project_uuid,omitempty" form:"project_uuid"`
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
	EventMarks []*struct {
		// 表ID
		ID string `json:"id,omitempty"`
		// 名稱
		Name string `json:"label,omitempty"`
		// 日期
		Day *time.Time `json:"day,omitempty"`
		// 專案UUID
		ProjectUUID string `json:"project_uuid,omitempty"`
		// 創建者
		CreatedBy string `json:"created_by,omitempty"`
		// 更新者
		UpdatedBy string `json:"updated_by,omitempty"`
		// 時間戳記
		section.TimeAt
	} `json:"event_marks"`
	// 分頁返回結構檔
	page.Total
}

// Single return structure file
type Single struct {
	// 表ID
	ID string `json:"id,omitempty"`
	// 名稱
	Name string `json:"label,omitempty"`
	// 日期
	Day *time.Time `json:"day,omitempty"`
	// 專案UUID
	ProjectUUID string `json:"project_uuid,omitempty"`
	// 創建者
	CreatedBy string `json:"created_by,omitempty"`
	// 更新者
	UpdatedBy string `json:"updated_by,omitempty"`
	// 時間戳記
	section.TimeAt
}

// Update struct is used end_date update achieves
type Update struct {
	// 表ID
	ID string `json:"id,omitempty"  binding:"omitempty,uuid4" validate:"omitempty,uuid4" swaggerignore:"true"`
	// 名稱
	Name *string `json:"label,omitempty"`
	// 日期
	Day *time.Time `json:"day,omitempty"`
	// 專案UUID
	ProjectUUID *string `json:"project_uuid,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4"`
	// 更新者
	UpdatedBy *string `json:"updated_by,omitempty" binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
}
