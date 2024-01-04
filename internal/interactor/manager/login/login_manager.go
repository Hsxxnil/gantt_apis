package login

import (
	"errors"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/ggwhite/go-masker"
	"gorm.io/gorm"
	"hta/config"
	jwxModel "hta/internal/interactor/models/jwx"
	loginModel "hta/internal/interactor/models/logins"
	roleModel "hta/internal/interactor/models/roles"
	userModel "hta/internal/interactor/models/users"
	"hta/internal/interactor/pkg/email"
	"hta/internal/interactor/pkg/jwx"
	"hta/internal/interactor/pkg/otp"
	"hta/internal/interactor/pkg/util"
	"hta/internal/interactor/pkg/util/code"
	"hta/internal/interactor/pkg/util/log"
	affiliationService "hta/internal/interactor/service/affiliation"
	jwxService "hta/internal/interactor/service/jwx"
	resourceService "hta/internal/interactor/service/resource"
	roleService "hta/internal/interactor/service/role"
	userService "hta/internal/interactor/service/user"
)

type Manager interface {
	Login(input *loginModel.Login) (int, any)
	Refresh(input *jwxModel.Refresh) (int, any)
	Verify(input *loginModel.Verify) (int, any)
	Forget(input *loginModel.Forget) (int, any)
	Register(input *loginModel.Register) (int, any)
}

type manager struct {
	UserService        userService.Service
	JwxService         jwxService.Service
	RoleService        roleService.Service
	ResourceService    resourceService.Service
	AffiliationService affiliationService.Service
}

func Init(db *gorm.DB) Manager {
	return &manager{
		UserService:        userService.Init(db),
		JwxService:         jwxService.Init(),
		RoleService:        roleService.Init(db),
		ResourceService:    resourceService.Init(db),
		AffiliationService: affiliationService.Init(db),
	}
}

func (m *manager) Login(input *loginModel.Login) (int, any) {
	var output any
	// verify username & password
	acknowledge, userBase, err := m.UserService.AcknowledgeUser(&userModel.Field{
		UserName: util.PointerString(input.UserName),
		Password: util.PointerString(input.Password),
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	if !acknowledge {
		return code.PermissionDenied, code.GetCodeMessage(code.PermissionDenied, "Incorrect username or password.")
	}

	// select authentication method
	switch input.ChangeTo {
	case 1:
		// generate passcode
		passcode, err := otp.GeneratePasscode(*userBase.OtpSecret)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}

		// send passcode to email
		to := *userBase.Email
		fromAddress := "REMOVED"
		fromName := "PMIS平台"
		mailPwd := "REMOVED"
		subject := "【PMIS平台】系統驗證碼(請勿回覆此郵件)"
		message := fmt.Sprintf(`
		<html lang="zh-TW">
		<head>
			<meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
			<style>
				body {
					font-family: 'Arial', sans-serif;
					background-color: #fff;
					color: #000;
				}
		
				.container {
					max-width: 450px;
					margin: 0 auto;
					padding: 20px;
					border-radius: 5px;
					border: 1px solid #cccccc;
					position: relative;
					color: #000;
				}
		
				.header {
					text-align: center;
					color: #000;
				}
		
		
				.footerMsg{
					font-size: small;
					color: #737171;
				}
		
				#passcodeContainer {
					text-align: center;
					background-color: #f1eeec;
					padding: 20px;
				}
		
				.passcode {
					font-size: 50px;
					color: #032942;
					letter-spacing: 10px;
					display: block;
					margin-bottom: 10px;
				}
		
				.expire {
					font-size: 15px;
					color: #737171;
					display: block;
				}
			</style>
		</head>
		<body>
		<div class="header">
			<h2>系統驗證碼</h2>
		</div>
		<div class="container">
			<p>親愛的用戶：</p>
			<p>感謝您使用PMIS專案管理平台，請輸入以下驗證碼。</p>
			<div id="passcodeContainer">
				<label class="passcode">%s</label>
				<label class="expire">時效為60秒</label>
			</div>
			<p>祝您使用愉快！</p>
			<p class="footerMsg">注意：此郵件由系統自動發出，請勿直接回覆。</p>
		</div>
		</body>
		</html>
		`, passcode)

		err = email.SendEmailWithHtml(to, fromAddress, fromName, mailPwd, subject, message)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}

		// mask email
		obscuredEmail := masker.Email(*userBase.Email)
		output = obscuredEmail

	case 2:
		output = "Please use authenticator to login."

	default:
		if !*userBase.IsAuthenticator {
			// generate passcode
			passcode, err := otp.GeneratePasscode(*userBase.OtpSecret)
			if err != nil {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}

			// send passcode to email
			to := *userBase.Email
			fromAddress := "REMOVED"
			fromName := "PMIS平台"
			mailPwd := "REMOVED"
			subject := "【PMIS平台】系統驗證碼(請勿回覆此郵件)"
			message := fmt.Sprintf(`
			<html lang="zh-TW">
			<head>
				<meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
				<style>
					body {
						font-family: 'Arial', sans-serif;
						background-color: #fff;
						color: #000;
					}
			
					.container {
						max-width: 450px;
						margin: 0 auto;
						padding: 20px;
						border-radius: 5px;
						border: 1px solid #cccccc;
						position: relative;
						color: #000;
					}
			
					.header {
						text-align: center;
						color: #000;
					}

					.footerMsg{
						font-size: small;
						color: #737171;
					}
			
					#passcodeContainer {
						text-align: center;
						background-color: #f1eeec;
						padding: 20px;
					}
			
					.passcode {
						font-size: 50px;
						color: #032942;
						letter-spacing: 10px;
						display: block;
						margin-bottom: 10px;
					}
			
					.expire {
						font-size: 15px;
						color: #737171;
						display: block;
					}
				</style>
			</head>
			<body>
			<div class="header">
				<h2>系統驗證碼</h2>
			</div>
			<div class="container">
				<p>親愛的用戶：</p>
				<p>感謝您使用PMIS專案管理平台，請輸入以下驗證碼。</p>
				<div id="passcodeContainer">
					<label class="passcode">%s</label>
					<label class="expire">時效為60秒</label>
				</div>
				<p>祝您使用愉快！</p>
				<p class="footerMsg">注意：此郵件由系統自動發出，請勿直接回覆。</p>
			</div>
			</body>
			</html>
			`, passcode)

			err = email.SendEmailWithHtml(to, fromAddress, fromName, mailPwd, subject, message)
			if err != nil {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}

			// mask email
			obscuredEmail := masker.Email(*userBase.Email)
			output = obscuredEmail
		} else {
			output = "Please use authenticator to login."
		}
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) Verify(input *loginModel.Verify) (int, any) {
	// get user
	userBase, err := m.UserService.GetBySingle(&userModel.Field{
		UserName: util.PointerString(input.UserName),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, "User does not exist.")
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// validate otp
	_, err = otp.ValidateOTP(input.Passcode, *userBase.OtpSecret)
	if err != nil {
		log.Error(err)
		return code.PermissionDenied, code.GetCodeMessage(code.PermissionDenied, "Incorrect passcode.")
	}

	// get role
	roleBase, err := m.RoleService.GetBySingle(&roleModel.Field{
		ID: *userBase.RoleID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, "Role does not exist.")
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// generate access token
	output := &jwxModel.Token{}
	accessToken, err := m.JwxService.CreateAccessToken(&jwxModel.JWX{
		UserID:     userBase.ID,
		Name:       userBase.Name,
		ResourceID: userBase.ResourceUUID,
		Role:       roleBase.Name,
	})

	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	accessTokenByte, _ := sonic.Marshal(accessToken)
	err = sonic.Unmarshal(accessTokenByte, &output)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// generate refresh token
	refreshToken, err := m.JwxService.CreateRefreshToken(&jwxModel.JWX{
		UserID: userBase.ID,
	})

	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	refreshTokenByte, _ := sonic.Marshal(refreshToken)
	err = sonic.Unmarshal(refreshTokenByte, &output)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// determine if the user has completed the information
	if userBase.ResourceUUID == nil {
		output.IsComplete = false
	} else {
		output.IsComplete = true
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) Refresh(input *jwxModel.Refresh) (int, any) {
	// verify refresh token
	j := &jwx.JWT{
		PublicKey: config.RefreshPublicKey,
		Token:     input.RefreshToken,
	}

	if len(input.RefreshToken) == 0 {
		return code.JWTRejected, code.GetCodeMessage(code.JWTRejected, "RefreshToken is null.")
	}

	j, err := j.Verify()
	if err != nil {
		log.Error(err)
		return code.JWTRejected, code.GetCodeMessage(code.JWTRejected, "RefreshToken is error.")
	}

	// get user
	field, err := m.UserService.GetBySingle(&userModel.Field{
		ID: j.Other["user_id"].(string),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.JWTRejected, code.GetCodeMessage(code.JWTRejected, "RefreshToken is error.")
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// get role
	roleBase, err := m.RoleService.GetBySingle(&roleModel.Field{
		ID: *field.RoleID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, "Role is not found.")
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// generate access token
	token, err := m.JwxService.CreateAccessToken(&jwxModel.JWX{
		UserID:     field.ID,
		Name:       field.Name,
		ResourceID: field.ResourceUUID,
		Role:       roleBase.Name,
	})
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	token.RefreshToken = input.RefreshToken
	return code.Successful, code.GetCodeMessage(code.Successful, token)
}

func (m *manager) Forget(input *loginModel.Forget) (int, any) {
	// get user by email
	userBase, err := m.UserService.GetBySingle(&userModel.Field{
		Email: util.PointerString(input.Email),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.BadRequest, code.GetCodeMessage(code.BadRequest, "User does not exist.")
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// get role
	roleBase, err := m.RoleService.GetBySingle(&roleModel.Field{
		ID: *userBase.RoleID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, "Role does not exist.")
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// generate access token
	accessToken, err := m.JwxService.CreateAccessToken(&jwxModel.JWX{
		UserID:     userBase.ID,
		Name:       userBase.Name,
		ResourceID: userBase.ResourceUUID,
		Role:       roleBase.Name,
		Expiration: util.PointerInt64(30),
	})

	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// send link to email
	to := input.Email
	fromAddress := "REMOVED"
	fromName := "PMIS平台"
	mailPwd := "REMOVED"
	subject := "【PMIS平台】請重設密碼(請勿回覆此郵件)"
	domain := input.Domain
	httpMod := "https"
	// modify localhost port and httpMod for testing
	if input.Domain == "localhost" {
		if input.Port == "" {
			return code.BadRequest, code.GetCodeMessage(code.BadRequest, "Invalid port.")
		}
		domain = fmt.Sprintf("%s:%s", input.Domain, input.Port)
		httpMod = "http"
	}
	resetPasswordLink := fmt.Sprintf("%s://%s/password_reset/%s", httpMod, domain, accessToken.AccessToken)
	message := fmt.Sprintf(`
    <html>
        <head>
            <style>
                body {
                    font-family: 'Arial', sans-serif;
                    text-align: center;
                    margin: 20px;
                }

                p {
                    margin-bottom: 10px;
                }

                a {
                    text-decoration: none;
                }

                button {
                    padding: 10px;
                    background-color: #4CAF50;
                    color: white;
                    border: none;
                    border-radius: 5px;
                    cursor: pointer;
                    text-decoration: none;
                }

                button:hover {
                    background-color: #45a049;
                }
            </style>
        </head>
        <body>
            <p>親愛的用戶：</p>
            <p>請點擊以下連結重設密碼：</p>
            <a href="%s">
                <button>
                    重設密碼
                </button>
            </a>
            <br>
            <p>祝您使用愉快！</p>
            <p><注意></p>
            <p>*此郵件由系統自動發出，請勿直接回覆。</p>
            <p>*此連結有效期限為30分鐘。</p>
        </body>
    </html>
`, resetPasswordLink)

	err = email.SendEmailWithHtml(to, fromAddress, fromName, mailPwd, subject, message)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// mask email
	obscuredEmail := masker.Email(*userBase.Email)

	return code.Successful, code.GetCodeMessage(code.Successful, obscuredEmail)
}

func (m *manager) Register(input *loginModel.Register) (int, any) {
	// determine if the username is duplicate
	quantity, _ := m.UserService.GetByQuantity(&userModel.Field{
		UserName: util.PointerString(input.UserName),
		Email:    util.PointerString(input.Email),
	})

	if quantity > 0 {
		log.Info("User already exists. UserName: ", input.UserName, "email: ", input.Email)
		return code.BadRequest, code.GetCodeMessage(code.BadRequest, "User already exists.")
	}

	// generate otp secret & otp auth url
	otpSecret, optAuthURL, err := otp.GenerateOTP("hta", input.UserName)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	userBase, err := m.UserService.Create(&userModel.Create{
		UserName:   input.UserName,
		Password:   input.Password,
		Email:      input.Email,
		RoleID:     input.RoleID,
		OtpSecret:  otpSecret,
		OtpAuthUrl: optAuthURL,
		CreatedBy:  input.CreatedBy,
	})
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// get role
	roleBase, err := m.RoleService.GetBySingle(&roleModel.Field{
		ID: *userBase.RoleID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, "Role does not exist.")
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// generate access token
	accessToken, err := m.JwxService.CreateAccessToken(&jwxModel.JWX{
		UserID:     userBase.ID,
		Name:       userBase.Name,
		ResourceID: userBase.ResourceUUID,
		Role:       roleBase.Name,
		Expiration: util.PointerInt64(30),
	})

	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// send link to email
	to := input.Email
	fromAddress := "REMOVED"
	fromName := "PMIS平台"
	mailPwd := "REMOVED"
	subject := "【PMIS平台】請驗證信箱以完成註冊(請勿回覆此郵件)"
	domain := input.Domain
	httpMod := "https"
	// modify localhost port and httpMod for testing
	if input.Domain == "localhost" {
		if input.Port == "" {
			return code.BadRequest, code.GetCodeMessage(code.BadRequest, "Invalid port.")
		}
		domain = fmt.Sprintf("%s:%s", input.Domain, input.Port)
		httpMod = "http"
	}
	resetPasswordLink := fmt.Sprintf("%s://%s/activate/%s", httpMod, domain, accessToken.AccessToken)
	message := fmt.Sprintf(`
    <html>
        <head>
            <style>
                body {
                    font-family: 'Arial', sans-serif;
                    text-align: center;
                    margin: 20px;
                }

                p {
                    margin-bottom: 10px;
                }

                a {
                    text-decoration: none;
                }

                button {
                    padding: 10px;
                    background-color: #4CAF50;
                    color: white;
                    border: none;
                    border-radius: 5px;
                    cursor: pointer;
                    text-decoration: none;
                }

                button:hover {
                    background-color: #45a049;
                }
            </style>
        </head>
        <body>
            <p>親愛的用戶：</p>
            <p>感謝您註冊本平台，請點擊以下連結驗證信箱：</p>
            <a href="%s">
                <button>
                    驗證信箱
                </button>
            </a>
            <br>
            <p>祝您使用愉快！</p>
            <p><注意></p>
            <p>*此郵件由系統自動發出，請勿直接回覆。</p>
            <p>*此連結有效期限為30分鐘。</p>
        </body>
    </html>
`, resetPasswordLink)

	err = email.SendEmailWithHtml(to, fromAddress, fromName, mailPwd, subject, message)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// mask email
	obscuredEmail := masker.Email(*userBase.Email)

	return code.Successful, code.GetCodeMessage(code.Successful, obscuredEmail)
}
