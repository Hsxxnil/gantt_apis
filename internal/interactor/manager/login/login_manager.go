package login

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ggwhite/go-masker"
	"gorm.io/gorm"
	"hta/config"
	jwxModel "hta/internal/interactor/models/jwx"
	loginModel "hta/internal/interactor/models/logins"
	organizationModel "hta/internal/interactor/models/organizations"
	roleModel "hta/internal/interactor/models/roles"
	userModel "hta/internal/interactor/models/users"
	"hta/internal/interactor/pkg/email"
	"hta/internal/interactor/pkg/jwx"
	"hta/internal/interactor/pkg/otp"
	"hta/internal/interactor/pkg/util"
	"hta/internal/interactor/pkg/util/code"
	"hta/internal/interactor/pkg/util/log"
	jwxService "hta/internal/interactor/service/jwx"
	organizationService "hta/internal/interactor/service/organization"
	roleService "hta/internal/interactor/service/role"
	userService "hta/internal/interactor/service/user"
)

type Manager interface {
	Login(input *loginModel.Login) (int, any)
	Refresh(input *jwxModel.Refresh) (int, any)
	Verify(input *loginModel.Verify) (int, any)
	Forget(input *loginModel.Forget) (int, any)
}

type manager struct {
	UserService         userService.Service
	JwxService          jwxService.Service
	RoleService         roleService.Service
	OrganizationService organizationService.Service
}

func Init(db *gorm.DB) Manager {
	return &manager{
		UserService:         userService.Init(db),
		JwxService:          jwxService.Init(),
		RoleService:         roleService.Init(db),
		OrganizationService: organizationService.Init(db),
	}
}

func (m *manager) Login(input *loginModel.Login) (int, any) {
	// get organization
	organizationBase, err := m.OrganizationService.GetBySingle(&organizationModel.Field{
		Domain: util.PointerString(input.Domain),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.BadRequest, code.GetCodeMessage(code.BadRequest, "Invalid domain.")
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// verify username & password
	acknowledge, userBase, err := m.UserService.AcknowledgeUser(&userModel.Field{
		UserName: util.PointerString(input.UserName),
		Password: util.PointerString(input.Password),
		OrgID:    util.PointerString(*organizationBase.ID),
	})
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	if !acknowledge {
		return code.PermissionDenied, code.GetCodeMessage(code.PermissionDenied, "Incorrect username or password.")
	}

	// generate otp secret & otp auth url
	// todo move to sign up
	otpSecret, optAuthURL, err := otp.GenerateOTP(*organizationBase.Name, input.UserName)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// generate passcode
	passcode, err := otp.GeneratePasscode(otpSecret)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// update otp secret & otp auth url
	err = m.UserService.Update(&userModel.Update{
		ID:         *userBase.ID,
		OtpSecret:  util.PointerString(otpSecret),
		OtpAuthUrl: util.PointerString(optAuthURL),
	})

	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// send passcode to email
	to := *userBase.Email
	fromAddress := "calla.nkust@gmail.com"
	fromName := "PMIS平台"
	mailPwd := "pfyj mkee hpgy sqlj"
	subject := "【PMIS平台】系統驗證碼(請勿回覆此郵件)"
	message := fmt.Sprintf(
		"親愛的用戶：\n"+
			"感謝您使用PMIS專案管理平台，請於30秒內輸入以下驗證碼。\n\n"+
			"驗證碼：%s\n"+
			"祝您使用愉快！\n\n"+
			"<注意>\n"+
			"*此郵件由系統自動發出，請勿直接回覆。", passcode)

	err = email.SendEmailWithText(to, fromAddress, fromName, mailPwd, subject, message)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// mask email
	obscuredEmail := masker.Email(*userBase.Email)

	return code.Successful, code.GetCodeMessage(code.Successful, obscuredEmail)
}

func (m *manager) Verify(input *loginModel.Verify) (int, any) {
	// get organization
	organizationBase, err := m.OrganizationService.GetBySingle(&organizationModel.Field{
		Domain: util.PointerString(input.Domain),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.BadRequest, code.GetCodeMessage(code.BadRequest, "Invalid domain.")
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// get user
	userBase, err := m.UserService.GetBySingle(&userModel.Field{
		UserName: util.PointerString(input.UserName),
		OrgID:    organizationBase.ID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, "User does not exist.")
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// validate otp
	otpValid, err := otp.ValidateOTP(input.Passcode, *userBase.OtpSecret)
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
		OrgID:      userBase.OrgID,
	})

	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	accessTokenByte, _ := json.Marshal(accessToken)
	err = json.Unmarshal(accessTokenByte, &output)
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

	refreshTokenByte, _ := json.Marshal(refreshToken)
	err = json.Unmarshal(refreshTokenByte, &output)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output.Role = *roleBase.Name
	output.UserID = *userBase.ID
	output.OtpVerified = otpValid
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
		OrgID:      field.OrgID,
	})
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	token.RefreshToken = input.RefreshToken
	token.Role = *roleBase.Name
	token.UserID = *field.ID
	return code.Successful, code.GetCodeMessage(code.Successful, token)
}

func (m *manager) Forget(input *loginModel.Forget) (int, any) {
	// get organization
	organizationBase, err := m.OrganizationService.GetBySingle(&organizationModel.Field{
		Domain: util.PointerString(input.Domain),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.BadRequest, code.GetCodeMessage(code.BadRequest, "Invalid domain.")
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// get user by email
	userBase, err := m.UserService.GetBySingle(&userModel.Field{
		Email: util.PointerString(input.Email),
		OrgID: organizationBase.ID,
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
		OrgID:      userBase.OrgID,
		Expiration: util.PointerInt64(30),
	})

	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// send link to email
	to := input.Email
	fromAddress := "calla.nkust@gmail.com"
	fromName := "PMIS平台"
	mailPwd := "pfyj mkee hpgy sqlj"
	subject := "【PMIS平台】請重設密碼(請勿回覆此郵件)"
	domain := input.Domain
	httpMod := "https"
	// modify localhost port and httpMod for testing
	if input.Domain == "localhost" {
		if input.Port != "" {
			domain = fmt.Sprintf("%s:%s", input.Domain, input.Port)
			httpMod = "http"
		} else {
			return code.BadRequest, code.GetCodeMessage(code.BadRequest, "Invalid port.")
		}
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
