package users

import (
	"hta/internal/interactor/models/page"
	"hta/internal/interactor/models/section"
)

// Create struct is used to create achieves
type Create struct {
	// 使用者名稱
	UserName string `json:"user_name,omitempty" binding:"required" validate:"required"`
	// 使用者中文名稱
	Name string `json:"name,omitempty" binding:"required" validate:"required"`
	// 資源UUID
	ResourceUUID *string `json:"resource_uuid,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4"`
	// 使用者密碼
	Password string `json:"password,omitempty" binding:"required" validate:"required"`
	// 使用者電話
	PhoneNumber string `json:"phone_number,omitempty"`
	// 使用者電子郵件
	Email string `json:"email,omitempty" binding:"required,email" validate:"required,email"`
	// 角色ID
	RoleID string `json:"role_id,omitempty" binding:"required,uuid4" validate:"required,uuid4"`
	// 公司ID
	CompanyID string `json:"company_id,omitempty" binding:"required,uuid4" validate:"required,uuid4"`
	// 創建者
	CreatedBy string `json:"created_by,omitempty" binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
}

// Field is structure file for search
type Field struct {
	// 表ID
	ID string `json:"id,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4" swaggerignore:"true"`
	// 使用者名稱
	UserName *string `json:"user_name,omitempty" form:"user_name"`
	// 使用者中文名稱
	Name *string `json:"name,omitempty" form:"name"`
	// 資源UUID
	ResourceUUID *string `json:"resource_uuid,omitempty" form:"resource_uuid"`
	// 使用者密碼
	Password *string `json:"password,omitempty" form:"password"`
	// 使用者電話
	PhoneNumber *string `json:"phone_number,omitempty" form:"phone_number"`
	// 使用者電子郵件
	Email *string `json:"email,omitempty" form:"email"`
	// 角色ID
	RoleID string `json:"role_id,omitempty" form:"role_id"`
	// 公司ID
	CompanyID *string `json:"company_id,omitempty" form:"company_id"`
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
	Users []*struct {
		// 表ID
		ID string `json:"id,omitempty"`
		// 使用者名稱
		UserName string `json:"user_name,omitempty"`
		// 使用者中文名稱
		Name string `json:"name,omitempty"`
		// 資源UUID
		ResourceUUID string `json:"resource_uuid,omitempty"`
		// 使用者電話
		PhoneNumber string `json:"phone_number,omitempty"`
		// 使用者電子郵件
		Email string `json:"email,omitempty"`
		// 角色ID
		RoleID string `json:"role_id,omitempty"`
		// 角色
		Role string `json:"role,omitempty"`
		// 創建者
		CreatedBy string `json:"created_by,omitempty"`
		// 更新者
		UpdatedBy string `json:"updated_by,omitempty"`
		// 時間戳記
		section.TimeAt
	} `json:"users"`
	// 分頁返回結構檔
	page.Total
}

// ListNoPagination is multiple return structure files without pagination
type ListNoPagination struct {
	// 多筆
	Users []*struct {
		// 表ID
		ID string `json:"id,omitempty"`
		// 使用者中文名稱
		Name string `json:"name,omitempty"`
	} `json:"users"`
}

// Single return structure file
type Single struct {
	// 表ID
	ID string `json:"id,omitempty"`
	// 使用者名稱
	UserName string `json:"user_name,omitempty"`
	// 使用者中文名稱
	Name string `json:"name,omitempty"`
	// 資源UUID
	ResourceUUID string `json:"resource_uuid,omitempty"`
	// 使用者電話
	PhoneNumber string `json:"phone_number,omitempty"`
	// 使用者電子郵件
	Email string `json:"email,omitempty"`
	// 角色ID
	RoleID string `json:"role_id,omitempty"`
	// 角色
	Role string `json:"role,omitempty"`
	// otp auth url
	OtpAuthUrl string `json:"otp_auth_url,omitempty"`
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
	ID string `json:"id,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4" swaggerignore:"true"`
	// 使用者名稱
	UserName *string `json:"user_name,omitempty"`
	// 使用者中文名稱
	Name *string `json:"name,omitempty"`
	// 資源UUID
	ResourceUUID *string `json:"resource_uuid,omitempty"`
	// 使用者密碼
	Password *string `json:"password,omitempty"`
	// 使用者舊密碼
	OldPassword *string `json:"old_password,omitempty"`
	// 使用者電話
	PhoneNumber *string `json:"phone_number,omitempty"`
	// 使用者電子郵件
	Email *string `json:"email,omitempty"`
	// 角色ID
	RoleID *string `json:"role_id,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4"`
	// otp secret
	OtpSecret *string `json:"otp_secret,omitempty"`
	// otp auth url
	OtpAuthUrl *string `json:"otp_auth_url,omitempty"`
	// 公司ID
	CompanyID *string `json:"company_id,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4"`
	// 更新者
	UpdatedBy *string `json:"updated_by,omitempty" binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
}

// ResetPassword struct is used to reset password
type ResetPassword struct {
	// 表ID
	ID string `json:"id,omitempty" binding:"omitempty,uuid4" validate:"omitempty,uuid4" swaggerignore:"true"`
	// 密碼
	Password string `json:"password,omitempty" binding:"required" validate:"required"`
	// 更新者
	UpdatedBy *string `json:"updated_by,omitempty" binding:"required,uuid4" validate:"required,uuid4" swaggerignore:"true"`
}
