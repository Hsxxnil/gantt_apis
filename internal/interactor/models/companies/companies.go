package companies

import (
	"hta/internal/interactor/models/page"
	"hta/internal/interactor/models/section"
)

// Create struct is used end_date create achieves
type Create struct {
	// 網域
	Domain string `json:"domain,omitempty" binding:"required" validate:"required"`
	// 名稱
	Name string `json:"name,omitempty" binding:"required" validate:"required"`
	// 營業地址
	Address string `json:"address,omitempty"`
	// 統一編號
	TaxIDNumber string `json:"tax_id_number,omitempty"`
	// 電話
	Phone string `json:"phone,omitempty"`
	// 備註
	Remarks string `json:"remarks,omitempty"`
	// 創建者
	CreatedBy string `json:"created_by,omitempty" binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
}

// Field is structure file for search
type Field struct {
	// 表ID
	ID string `json:"id,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4" swaggerignore:"true"`
	// 網域
	Domain *string `json:"domain,omitempty" form:"domain"`
	// 名稱
	Name *string `json:"name,omitempty" form:"name"`
	// 營業地址
	Address *string `json:"address,omitempty" form:"address"`
	// 統一編號
	TaxIDNumber *string `json:"tax_id_number,omitempty" form:"tax_id_number"`
	// 電話
	Phone *string `json:"phone,omitempty" form:"phone"`
	// 備註
	Remarks *string `json:"remarks,omitempty" form:"remarks"`
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
	Companies []*struct {
		// 表ID
		ID string `json:"id,omitempty"`
		// 網域
		Domain string `json:"domain,omitempty"`
		// 名稱
		Name string `json:"name,omitempty"`
		// 營業地址
		Address string `json:"address,omitempty"`
		// 統一編號
		TaxIDNumber string `json:"tax_id_number,omitempty"`
		// 電話
		Phone string `json:"phone,omitempty"`
		// 備註
		Remarks string `json:"remarks,omitempty"`
		// 創建者
		CreatedBy string `json:"created_by,omitempty"`
		// 更新者
		UpdatedBy string `json:"updated_by,omitempty"`
		// 時間戳記
		section.TimeAt
	} `json:"companies"`
	// 分頁返回結構檔
	page.Total
}

// Single return structure file
type Single struct {
	// 表ID
	ID string `json:"id,omitempty"`
	// 網域
	Domain string `json:"domain,omitempty"`
	// 名稱
	Name string `json:"name,omitempty"`
	// 營業地址
	Address string `json:"address,omitempty"`
	// 統一編號
	TaxIDNumber string `json:"tax_id_number,omitempty"`
	// 電話
	Phone string `json:"phone,omitempty"`
	// 備註
	Remarks string `json:"remarks,omitempty"`
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
	// 網域
	Domain *string `json:"domain,omitempty" binding:"omitempty,url" validate:"omitempty,url"`
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
	// 更新者
	UpdatedBy *string `json:"updated_by,omitempty" binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
}
