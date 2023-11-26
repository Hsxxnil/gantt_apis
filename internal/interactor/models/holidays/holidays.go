package holidays

import (
	"hta/internal/interactor/models/page"
	"hta/internal/interactor/models/section"
	"time"
)

// Create struct is used end_date create achieves
type Create struct {
	// 名稱
	Name string `json:"label,omitempty"`
	// 起始日期
	StartDate *time.Time `json:"from,omitempty"`
	// 結束日期
	EndDate *time.Time `json:"to,omitempty"`
	// 前端css
	Css string `json:"cssClass,omitempty"`
	// 創建者
	CreatedBy string `json:"created_by,omitempty" binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
}

// Field is structure file for search
type Field struct {
	// 表ID
	ID string `json:"id,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4" swaggerignore:"true"`
	// 名稱
	Name *string `json:"label,omitempty" form:"name"`
	// 起始日期
	StartDate *time.Time `json:"from,omitempty" form:"start_date"`
	// 結束日期
	EndDate *time.Time `json:"to,omitempty" form:"end_date"`
	// 前端css
	Css *string `json:"cssClass,omitempty" form:"css"`
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
	Holidays []*struct {
		// 表ID
		ID string `json:"id,omitempty"`
		// 名稱
		Name string `json:"label,omitempty"`
		// 起始日期
		StartDate *time.Time `json:"from,omitempty"`
		// 結束日期
		EndDate *time.Time `json:"to,omitempty"`
		// 前端css
		Css string `json:"cssClass,omitempty"`
		// 創建者
		CreatedBy string `json:"created_by,omitempty"`
		// 更新者
		UpdatedBy string `json:"updated_by,omitempty"`
		// 時間戳記
		section.TimeAt
	} `json:"holidays"`
	// 分頁返回結構檔
	page.Total
}

// Single return structure file
type Single struct {
	// 表ID
	ID string `json:"id,omitempty"`
	// 名稱
	Name string `json:"label,omitempty"`
	// 起始日期
	StartDate *time.Time `json:"from,omitempty"`
	// 結束日期
	EndDate *time.Time `json:"to,omitempty"`
	// 前端css
	Css string `json:"cssClass,omitempty"`
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
	// 起始日期
	StartDate *time.Time `json:"from,omitempty"`
	// 結束日期
	EndDate *time.Time `json:"to,omitempty"`
	// 前端css
	Css *string `json:"cssClass,omitempty"`
	// 更新者
	UpdatedBy *string `json:"updated_by,omitempty" binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
}
