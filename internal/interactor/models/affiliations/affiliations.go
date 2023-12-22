package affiliations

import (
	"hta/internal/interactor/models/page"
)

// Create struct is used end_date create achieves
type Create struct {
	// 使用者ID
	UserID string `json:"user_id,omitempty" binding:"required,uuid4" validate:"required,uuid4"`
	// 部門ID
	DeptID string `json:"dept_id,omitempty" binding:"required,uuid4" validate:"required,uuid4"`
	// 職稱
	JobTitle string `json:"job_title,omitempty"`
	// 是否為主管
	IsSupervisor bool `json:"is_supervisor,omitempty" binding:"required" validate:"required"`
	// 創建者
	CreatedBy string `json:"created_by,omitempty" binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
}

// Field is structure file for search
type Field struct {
	// 表ID
	ID string `json:"id,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4" swaggerignore:"true"`
	// 使用者ID
	UserID *string `json:"user_id,omitempty" form:"user_id"`
	// 部門ID
	DeptID *string `json:"dept_id,omitempty" form:"dept_id"`
	// 職稱
	JobTitle *string `json:"job_title,omitempty" form:"job_title"`
	// 是否為主管
	IsSupervisor *bool `json:"is_supervisor,omitempty" form:"is_supervisor"`
}

// Fields is the searched structure file (including pagination)
type Fields struct {
	// 搜尋結構檔
	Field
	// 分頁搜尋結構檔
	page.Pagination
}

// Update struct is used end_date update achieves
type Update struct {
	// 表ID
	ID string `json:"id,omitempty"  binding:"omitempty,uuid4" validate:"omitempty,uuid4" swaggerignore:"true"`
	// 使用者ID
	UserID *string `json:"user_id,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4"`
	// 部門ID
	DeptID *string `json:"dept_id,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4"`
	// 職稱
	JobTitle *string `json:"job_title,omitempty"`
	// 是否為主管
	IsSupervisor *bool `json:"is_supervisor,omitempty"`
	// 更新者
	UpdatedBy *string `json:"updated_by,omitempty" binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
}
