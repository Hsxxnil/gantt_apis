package login

import (
	"net/http"

	"hta/internal/interactor/manager/login"
	jwxModel "hta/internal/interactor/models/jwx"
	loginModel "hta/internal/interactor/models/logins"
	"hta/internal/interactor/pkg/util/code"
	"hta/internal/interactor/pkg/util/log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Control interface {
	Login(ctx *gin.Context)
	Verify(ctx *gin.Context)
	Refresh(ctx *gin.Context)
	Forget(ctx *gin.Context)
	Register(ctx *gin.Context)
}

type control struct {
	Manager login.Manager
}

func Init(db *gorm.DB) Control {
	return &control{
		Manager: login.Init(db),
	}
}

// Login
// @Summary 登入
// @description 登入
// @Tags login
// @version 1.0
// @Accept json
// @produce json
// @param * body logins.Login true "登入帶入"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /login [post]
func (c *control) Login(ctx *gin.Context) {
	input := &loginModel.Login{}

	if err := ctx.ShouldBindJSON(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.Login(input)
	ctx.JSON(httpCode, codeMessage)
}

// Verify
// @Summary 驗證
// @description 驗證
// @Tags login
// @version 1.0
// @Accept json
// @produce json
// @param * body logins.Verify true "驗證帶入"
// @success 200 object code.SuccessfulMessage{body=jwx.Token} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /verify [post]
func (c *control) Verify(ctx *gin.Context) {
	input := &loginModel.Verify{}

	if err := ctx.ShouldBindJSON(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.Verify(input)
	ctx.JSON(httpCode, codeMessage)
}

// Refresh
// @Summary 換新的令牌
// @description 換新的令牌
// @Tags login
// @version 1.0
// @Accept json
// @produce json
// @param * body jwx.Refresh true "登入帶入"
// @success 200 object code.SuccessfulMessage{body=jwx.Token} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /refresh [post]
func (c *control) Refresh(ctx *gin.Context) {
	input := &jwxModel.Refresh{}
	if err := ctx.ShouldBindJSON(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.Refresh(input)
	ctx.JSON(httpCode, codeMessage)
}

// Forget
// @Summary 忘記密碼
// @description 忘記密碼
// @Tags login
// @version 1.0
// @Accept json
// @produce json
// @param * body logins.Forget true "登入帶入"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /forget-password [post]
func (c *control) Forget(ctx *gin.Context) {
	input := &loginModel.Forget{}

	if err := ctx.ShouldBindJSON(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.Forget(input)
	ctx.JSON(httpCode, codeMessage)
}

// Register
// @Summary 註冊
// @description 註冊
// @Tags login
// @version 1.0
// @Accept json
// @produce json
// @param * body logins.Register true "註冊"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /register [post]
func (c *control) Register(ctx *gin.Context) {
	input := &loginModel.Register{}
	// set default value (created by admin & role id is user)
	input.CreatedBy = "7c0595cf-2d9a-4e77-858c-a33f9d1e8452"
	input.RoleID = "bcf2a32f-e801-4ae8-bc03-07c593ce626f"
	if err := ctx.ShouldBindJSON(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.Register(input)
	ctx.JSON(httpCode, codeMessage)
}
