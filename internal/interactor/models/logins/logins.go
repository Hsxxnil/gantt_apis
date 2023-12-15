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

// Forget struct is used to forget password.
type Forget struct {
	// 使用者電子郵件
	Email string `json:"email,omitempty" binding:"required,email" validate:"required,email"`
	// 網域
	Domain string `json:"domain,omitempty" binding:"required" validate:"required"`
	// 連接埠
	Port string `json:"port,omitempty"`
}
