package logins

// Login struct is used to log in
type Login struct {
	// 使用者名稱
	UserName string `json:"user_name,omitempty" binding:"required" validate:"required"`
	// 密碼
	Password string `json:"password,omitempty" binding:"required" validate:"required"`
}
