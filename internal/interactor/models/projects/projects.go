package projects

import (
	"hta/internal/interactor/models/page"
	"hta/internal/interactor/models/section"
	"time"
)

// Create struct is used to create achieves
type Create struct {
	// 名稱
	ProjectName string `json:"project_name,omitempty"`
	// 類別
	Type string `json:"type,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4"`
	// 代號
	Code string `json:"code,omitempty"`
	// 負責人
	Manager string `json:"manager,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4"`
	// 起始日期
	StartDate *time.Time `json:"start_date,omitempty"`
	// 結束日期
	EndDate *time.Time `json:"end_date,omitempty"`
	// 客戶
	Client string `json:"client,omitempty"`
	// 狀態
	Status string `json:"status,omitempty"`
	//資源
	Resource []*ProjectResource `json:"resource,omitempty"`
	// 創建者
	CreatedBy string `json:"created_by,omitempty" binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
}

// Field is structure file for search
type Field struct {
	// 表ID
	ProjectUUID string `json:"project_uuid,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4" swaggerignore:"true"`
	// 名稱
	ProjectName *string `json:"project_name,omitempty" form:"project_name"`
	// 類別
	Type *string `json:"type,omitempty" form:"type"`
	// 代號
	Code *string `json:"code,omitempty" form:"code"`
	// 負責人
	Manager *string `json:"manager,omitempty" form:"manager"`
	// 起始日期
	StartDate *time.Time `json:"start_date,omitempty" form:"start_date"`
	// 結束日期
	EndDate *time.Time `json:"end_date,omitempty" form:"end_date"`
	// 客戶
	Client *string `json:"client,omitempty" form:"client"`
	// 狀態
	Status *string `json:"status,omitempty" form:"status"`
	// 創建者
	CreatedBy *string `json:"created_by,omitempty" form:"created_by"`
	// 專案UUIDs (後端查詢用）
	ProjectIDs []*string `json:"project_ids,omitempty" form:"project_ids" swaggerignore:"true"`
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
	// 類別
	FilterType string `json:"type,omitempty"`
	// 類別 ids (後端查詢用）
	FilterTypes []string `json:"types,omitempty" swaggerignore:"true"`
	// 客戶
	FilterClient string `json:"client,omitempty"`
	// 名稱
	FilterName string `json:"project_name,omitempty"`
	// 負責人
	FilterManager string `json:"manager,omitempty"`
	// manager ids (後端查詢用）
	FilterManagers []string `json:"managers,omitempty" swaggerignore:"true"`
	// 代號
	FilterCode string `json:"code,omitempty"`
	// 起始日期
	FilterStartDate *time.Time `json:"start_date,omitempty"`
	// 結束日期
	FilterEndDate *time.Time `json:"end_date,omitempty"`
	// 狀態
	FilterStatus string `json:"status,omitempty"`
}

// List is multiple return structure files
type List struct {
	// 多筆
	Projects []*struct {
		// 表ID
		ProjectUUID string `json:"project_uuid,omitempty"`
		// 前端編號 (非表ID)
		ProjectID string `json:"project_id,omitempty"`
		// 名稱
		ProjectName string `json:"project_name,omitempty"`
		// 類別
		Type string `json:"type,omitempty"`
		// 代號
		Code string `json:"code,omitempty"`
		// 負責人
		Manager string `json:"manager,omitempty"`
		// 起始日期
		StartDate *time.Time `json:"start_date,omitempty"`
		// 結束日期
		EndDate *time.Time `json:"end_date,omitempty"`
		// 客戶
		Client string `json:"client,omitempty"`
		// 狀態
		Status string `json:"status,omitempty"`
		// 創建者
		CreatedBy string `json:"created_by,omitempty"`
		// 更新者
		UpdatedBy string `json:"updated_by,omitempty"`
		// 時間戳記
		section.TimeAt
	} `json:"projects"`
	// 分頁返回結構檔
	page.Total
}

// Single return structure file
type Single struct {
	// 表ID
	ProjectUUID string `json:"project_uuid,omitempty"`
	// 前端編號 (非表ID)
	ProjectID string `json:"project_id,omitempty"`
	// 名稱
	ProjectName string `json:"project_name,omitempty"`
	// 類別
	Type string `json:"type,omitempty"`
	// 代號
	Code string `json:"code,omitempty"`
	// 負責人
	Manager string `json:"manager,omitempty"`
	// 起始日期
	StartDate *time.Time `json:"start_date,omitempty"`
	// 結束日期
	EndDate *time.Time `json:"end_date,omitempty"`
	// 客戶
	Client string `json:"client,omitempty"`
	// 狀態
	Status string `json:"status,omitempty"`
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
	ProjectUUID string `json:"project_uuid,omitempty"  binding:"omitempty,uuid4" validate:"omitempty,uuid4" swaggerignore:"true"`
	// 名稱
	ProjectName *string `json:"project_name,omitempty"`
	// 類別
	Type *string `json:"type,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4"`
	// 代號
	Code *string `json:"code,omitempty"`
	// 負責人
	Manager *string `json:"manager,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4"`
	// 起始日期
	StartDate *time.Time `json:"start_date,omitempty"`
	// 結束日期
	EndDate *time.Time `json:"end_date,omitempty"`
	// 客戶
	Client *string `json:"client,omitempty"`
	// 狀態
	Status *string `json:"status,omitempty"`
	//資源
	Resource []*ProjectResource `json:"resource,omitempty"`
	// 更新者
	UpdatedBy *string `json:"updated_by,omitempty" binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
}

// ProjectResource is used to sync create or update project_resource.
type ProjectResource struct {
	// 資源UUID
	ResourceUUID string `json:"resource_uuid,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4"`
	// 專案角色
	Role string `json:"role,omitempty"`
}
