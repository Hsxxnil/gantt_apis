package login

import (
	"encoding/json"
	"errors"
	"fmt"
	"hta/config"
	jwxModel "hta/internal/interactor/models/jwx"
	loginsModel "hta/internal/interactor/models/logins"
	roleModel "hta/internal/interactor/models/roles"
	usersModel "hta/internal/interactor/models/users"
	"hta/internal/interactor/pkg/jwx"
	"hta/internal/interactor/pkg/otp"
	"hta/internal/interactor/pkg/util"
	"hta/internal/interactor/pkg/util/code"
	"hta/internal/interactor/pkg/util/email"
	"hta/internal/interactor/pkg/util/log"
	jwxService "hta/internal/interactor/service/jwx"
	roleService "hta/internal/interactor/service/role"
	userService "hta/internal/interactor/service/user"

	"gorm.io/gorm"
)

type Manager interface {
	Login(input *loginsModel.Login) (int, any)
	Refresh(input *jwxModel.Refresh) (int, any)
	Verify(input *loginsModel.Verify) (int, any)
}

type manager struct {
	UserService userService.Service
	JwxService  jwxService.Service
	RoleService roleService.Service
}

func Init(db *gorm.DB) Manager {
	return &manager{
		UserService: userService.Init(db),
		JwxService:  jwxService.Init(),
		RoleService: roleService.Init(db),
	}
}

func (m *manager) Login(input *loginsModel.Login) (int, any) {
	// 驗證帳密
	acknowledge, userBase, err := m.UserService.AcknowledgeUser(&usersModel.Field{
		UserName: util.PointerString(input.UserName),
		Password: util.PointerString(input.Password),
	})
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	if acknowledge == false {
		return code.PermissionDenied, code.GetCodeMessage(code.PermissionDenied, "Incorrect username or password.")
	}

	// 產生otp secret & otp auth url
	otpSecret, optAuthURL, err := otp.GenerateOTP("sien", input.UserName)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// 產生驗證碼
	passcode, err := otp.GeneratePasscode(otpSecret)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// 更新otp secret & otp auth url
	err = m.UserService.Update(&usersModel.Update{
		ID:         *userBase.ID,
		OtpSecret:  util.PointerString(otpSecret),
		OtpAuthUrl: util.PointerString(optAuthURL),
	})

	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// 發送驗證碼
	to := *userBase.Email
	fromAddress := "REMOVED"
	fromName := "PMIS平台"
	mailPwd := "REMOVED"
	subject := "系統驗證碼(請勿回覆此郵件)"
	message := fmt.Sprintf(
		"您好：\n"+
			"歡迎使用PMIS專案管理平台，\n"+
			"您的驗證碼是：%s\n"+
			"敬祝愉快\n"+
			"此為系統發出的電子郵件，請勿直接回覆。", passcode)

	err = email.SendEmail(to, fromAddress, fromName, mailPwd, subject, message)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, "send passcode successful.")
}

func (m *manager) Verify(input *loginsModel.Verify) (int, any) {
	// 取得使用者
	userBase, err := m.UserService.GetBySingle(&usersModel.Field{
		UserName: util.PointerString(input.UserName),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, "User does not exist.")
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// 取得角色
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

	// 產生accessToken
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

	accessTokenByte, _ := json.Marshal(accessToken)
	err = json.Unmarshal(accessTokenByte, &output)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// 產生refreshToken
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
	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) Refresh(input *jwxModel.Refresh) (int, any) {
	// 驗證refreshToken
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

	field, err := m.UserService.GetBySingle(&usersModel.Field{
		ID: j.Other["user_id"].(string),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.JWTRejected, code.GetCodeMessage(code.JWTRejected, "RefreshToken is error.")
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// 取得角色
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

	// 產生accessToken
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
	token.Role = *roleBase.Name
	token.UserID = *field.ID
	return code.Successful, code.GetCodeMessage(code.Successful, token)
}
