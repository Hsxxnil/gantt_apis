package logins

// Login struct is used to log in
type Login struct {
	// 使用者名稱
	UserName string `json:"user_name,omitempty" binding:"required" validate:"required"`
	// 密碼
	Password string `json:"password,omitempty" binding:"required" validate:"required"`
	// 網域
	Domain string `json:"domain,omitempty" binding:"required" validate:"required"`
}

// Verify struct is used to verify the OTP code
type Verify struct {
	// 使用者名稱
	UserName string `json:"user_name,omitempty" binding:"required" validate:"required"`
	// 驗證碼
	Passcode string `json:"passcode,omitempty" binding:"required" validate:"required"`
	// 網域
	Domain string `json:"domain,omitempty" binding:"required" validate:"required"`
}
