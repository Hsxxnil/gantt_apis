package users

import (
	"hta/internal/entity/postgresql/db/roles"
	"hta/internal/interactor/models/special"
)

// Table struct is users database table struct
type Table struct {
	// 表ID
	ID string `gorm:"<-:create;column:id;type:uuid;not null;primaryKey;" json:"id"`
	// 使用者名稱
	UserName string `gorm:"column:user_name;type:text;not null;" json:"user_name"`
	// 使用者中文名稱
	Name string `gorm:"column:name;type:text;not null;" json:"name"`
	// 資源UUID
	ResourceUUID *string `gorm:"column:resource_uuid;type:text;" json:"resource_uuid"`
	// 使用者密碼
	Password string `gorm:"column:password;type:text;not null;" json:"password"`
	// 使用者電話
	PhoneNumber string `gorm:"column:phone_number;type:text;" json:"phone_number"`
	// 使用者電子郵件
	Email string `gorm:"column:email;type:text;" json:"email"`
	// 角色ID
	RoleID string `gorm:"column:role_id;type:uuid;not null;" json:"role_id"`
	// roles data
	Roles roles.Table `gorm:"foreignKey:RoleID;references:ID" json:"roles,omitempty"`
	// otp secret
	OtpSecret string `gorm:"column:otp_secret;type:text;" json:"otp_secret"`
	// otp auth url
	OtpAuthUrl string `gorm:"column:otp_auth_url;type:text;" json:"otp_auth_url"`
	// 公司ID
	CompanyID string `gorm:"column:company_id;type:uuid;" json:"company_id"`
	// 引入後端專用
	special.Table
}

// Base struct is corresponding to users table structure file
type Base struct {
	// 表ID
	ID *string `json:"id,omitempty"`
	// 使用者名稱
	UserName *string `json:"user_name,omitempty"`
	// 使用者中文名稱
	Name *string `json:"name,omitempty"`
	// 資源UUID
	ResourceUUID *string `json:"resource_uuid,omitempty"`
	// 使用者密碼
	Password *string `json:"password,omitempty"`
	// 使用者電話
	PhoneNumber *string `json:"phone_number,omitempty"`
	// 使用者電子郵件
	Email *string `json:"email,omitempty"`
	// 角色ID
	RoleID *string `json:"role_id,omitempty"`
	// roles data
	Roles roles.Base `json:"roles,omitempty"`
	// otp secret
	OtpSecret *string `json:"otp_secret,omitempty"`
	// otp auth url
	OtpAuthUrl *string `json:"otp_auth_url,omitempty"`
	// 公司ID
	CompanyID *string `json:"company_id,omitempty"`
	// 引入後端專用
	special.Base
}

// TableName sets the insert table name for this struct type
func (t *Table) TableName() string {
	return "users"
}
