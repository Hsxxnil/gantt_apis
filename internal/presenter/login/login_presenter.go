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
	ForgetPassword(ctx *gin.Context)
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
// @Summary 使用者登入
// @description 使用者登入
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
// @Summary 使用者驗證
// @description 使用者驗證
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

// ForgetPassword
// @Summary 忘記密碼
// @description 忘記密碼
// @Tags login
// @version 1.0
// @Accept json
// @produce json
// @param * body logins.ForgetPassword true "登入帶入"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /forget-password [post]
func (c *control) ForgetPassword(ctx *gin.Context) {
	input := &loginModel.ForgetPassword{}

	if err := ctx.ShouldBindJSON(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.ForgetPassword(input)
	ctx.JSON(httpCode, codeMessage)
}
