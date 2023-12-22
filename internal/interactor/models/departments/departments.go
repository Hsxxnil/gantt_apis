package departments

import (
	"hta/internal/interactor/models/page"
	"hta/internal/interactor/models/section"
)

// Create struct is used end_date create achieves
type Create struct {
	// 部門主管ID(user_id)
	SupervisorID *string `json:"supervisor_id,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4"`
	// 名稱
	Name string `json:"name,omitempty" binding:"required" validate:"required"`
	// 傳真
	Fax string `json:"fax,omitempty"`
	// 電話
	Tel string `json:"tel,omitempty"`
	// 創建者
	CreatedBy string `json:"created_by,omitempty" binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
}

// Field is structure file for search
type Field struct {
	// 表ID
	ID string `json:"id,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4" swaggerignore:"true"`
	// 部門主管ID(user_id)
	SupervisorID *string `json:"supervisor_id,omitempty" form:"supervisor_id"`
	// 名稱
	Name *string `json:"name,omitempty" form:"name"`
	// 傳真
	Fax *string `json:"fax,omitempty" form:"fax"`
	// 電話
	Tel *string `json:"tel,omitempty" form:"tel"`
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
	Departments []*struct {
		// 表ID
		ID string `json:"id,omitempty"`
		// 名稱
		Name string `json:"name,omitempty"`
		// 傳真
		Fax string `json:"fax,omitempty"`
		// 電話
		Tel string `json:"tel,omitempty"`
		// 創建者
		CreatedBy string `json:"created_by,omitempty"`
		// 更新者
		UpdatedBy string `json:"updated_by,omitempty"`
		// 時間戳記
		section.TimeAt
	} `json:"departments"`
	// 分頁返回結構檔
	page.Total
}

// Single return structure file
type Single struct {
	// 表ID
	ID string `json:"id,omitempty"`
	// 名稱
	Name string `json:"name,omitempty"`
	// 傳真
	Fax string `json:"fax,omitempty"`
	// 電話
	Tel string `json:"tel,omitempty"`
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
	// 部門主管ID(user_id)
	SupervisorID *string `json:"supervisor_id,omitempty" binding:"omitempty,url" validate:"omitempty,url"`
	// 名稱
	Name *string `json:"name,omitempty"`
	// 傳真
	Fax *string `json:"fax,omitempty"`
	// 電話
	Tel *string `json:"tel,omitempty"`
	// 更新者
	UpdatedBy *string `json:"updated_by,omitempty" binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
}
