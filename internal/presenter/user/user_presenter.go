package user

import (
	"net/http"

	"gantt/internal/interactor/pkg/util"

	constant "gantt/internal/interactor/constants"

	"gantt/internal/interactor/manager/user"
	userModel "gantt/internal/interactor/models/users"
	"gantt/internal/interactor/pkg/util/code"
	"gantt/internal/interactor/pkg/util/log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Control interface {
	GetByList(ctx *gin.Context)
	GetByListNoPagination(ctx *gin.Context)
	GetBySingle(ctx *gin.Context)
	GetByCurrent(ctx *gin.Context)
	Delete(ctx *gin.Context)
	Update(ctx *gin.Context)
	UpdateByCurrent(ctx *gin.Context)
	EnableByCurrent(ctx *gin.Context)
	ResetPassword(ctx *gin.Context)
	Duplicate(ctx *gin.Context)
	EnableAuthenticator(ctx *gin.Context)
	ChangeEmail(ctx *gin.Context)
	VerifyEmail(ctx *gin.Context)
}

type control struct {
	Manager user.Manager
}

func Init(db *gorm.DB) Control {
	return &control{
		Manager: user.Init(db),
	}
}

// GetByList
// @Summary 取得全部使用者
// @description 取得全部使用者
// @Tags user
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string  true "JWE Token"
// @param page query int true "目前頁數,請從1開始帶入"
// @param limit query int true "一次回傳比數,請從1開始帶入,最高上限20"
// @success 200 object code.SuccessfulMessage{body=users.List} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /users/list [post]
func (c *control) GetByList(ctx *gin.Context) {
	input := &userModel.Fields{}
	if err := ctx.ShouldBindQuery(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	if input.Limit >= constant.DefaultLimit {
		input.Limit = constant.DefaultLimit
	}

	httpCode, codeMessage := c.Manager.GetByList(input)
	ctx.JSON(httpCode, codeMessage)
}

// GetByListNoPagination
// @Summary 取得全部使用者(不用page和limit)
// @description 取得全部使用者(不用page和limit)
// @Tags user
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string  true "JWE Token"
// @success 200 object code.SuccessfulMessage{body=users.ListNoPagination} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /users [get]
func (c *control) GetByListNoPagination(ctx *gin.Context) {
	input := &userModel.Field{}
	if err := ctx.ShouldBindQuery(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.GetByListNoPagination(input)
	ctx.JSON(httpCode, codeMessage)
}

// GetBySingle
// @Summary 取得單一使用者
// @description 取得單一使用者
// @Tags user
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string  true "JWE Token"
// @param id path string true "使用者ID"
// @success 200 object code.SuccessfulMessage{body=users.Single} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /users/{id} [get]
func (c *control) GetBySingle(ctx *gin.Context) {
	id := ctx.Param("id")
	input := &userModel.Field{}
	input.ID = id
	if err := ctx.ShouldBindQuery(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.GetBySingle(input)
	ctx.JSON(httpCode, codeMessage)
}

// GetByCurrent
// @Summary 取得當前使用者
// @description 取得當前使用者
// @Tags user
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string  true "JWE Token"
// @success 200 object code.SuccessfulMessage{body=users.Single} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /users/current-user [get]
func (c *control) GetByCurrent(ctx *gin.Context) {
	input := &userModel.Field{}
	input.ID = ctx.MustGet("user_id").(string)
	if err := ctx.ShouldBindQuery(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.GetBySingle(input)
	ctx.JSON(httpCode, codeMessage)
}

// Delete
// @Summary 刪除單一使用者
// @description 刪除單一使用者
// @Tags user
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string  true "JWE Token"
// @param id path string true "使用者ID"
// @param * body users.Update true "更新使用者"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /users/{id} [delete]
func (c *control) Delete(ctx *gin.Context) {
	trx := ctx.MustGet("db_trx").(*gorm.DB)
	id := ctx.Param("id")
	input := &userModel.Update{}
	input.ID = id
	input.UpdatedBy = util.PointerString(ctx.MustGet("user_id").(string))
	if err := ctx.ShouldBindQuery(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.Delete(trx, input)
	ctx.JSON(httpCode, codeMessage)
}

// Update
// @Summary 更新使用者
// @description 更新使用者
// @Tags user
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string  true "JWE Token"
// @param id path string true "使用者ID"
// @param * body users.Update true "更新使用者"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /users/{id} [patch]
func (c *control) Update(ctx *gin.Context) {
	trx := ctx.MustGet("db_trx").(*gorm.DB)
	id := ctx.Param("id")
	input := &userModel.Update{}
	input.ID = id
	input.UpdatedBy = util.PointerString(ctx.MustGet("user_id").(string))
	if err := ctx.ShouldBindJSON(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.Update(trx, input)
	ctx.JSON(httpCode, codeMessage)
}

// UpdateByCurrent
// @Summary 更新當前使用者
// @description 更新當前使用者
// @Tags user
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string  true "JWE Token"
// @param * body users.Update true "更新使用者"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /users/current-user [patch]
func (c *control) UpdateByCurrent(ctx *gin.Context) {
	trx := ctx.MustGet("db_trx").(*gorm.DB)
	input := &userModel.Update{}
	input.ID = ctx.MustGet("user_id").(string)
	input.UpdatedBy = util.PointerString(ctx.MustGet("user_id").(string))
	if err := ctx.ShouldBindJSON(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.Update(trx, input)
	ctx.JSON(httpCode, codeMessage)
}

// EnableByCurrent
// @Summary 啟用當前使用者
// @description 啟用當前使用者
// @Tags user
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string  true "JWE Token"
// @param * body users.Enable true "啟用當前使用者"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /users/enable/current-user [post]
func (c *control) EnableByCurrent(ctx *gin.Context) {
	input := &userModel.Enable{}
	input.ID = ctx.MustGet("user_id").(string)
	input.UpdatedBy = util.PointerString(ctx.MustGet("user_id").(string))
	if err := ctx.ShouldBindQuery(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.Enable(input)
	ctx.JSON(httpCode, codeMessage)
}

// ResetPassword
// @Summary 重設密碼
// @description 重設密碼
// @Tags user
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string  true "JWE Token"
// @param * body users.ResetPassword true "重設密碼"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /users/reset-password/current-user [post]
func (c *control) ResetPassword(ctx *gin.Context) {
	input := &userModel.ResetPassword{}
	input.ID = ctx.MustGet("user_id").(string)
	input.UpdatedBy = util.PointerString(ctx.MustGet("user_id").(string))
	if err := ctx.ShouldBindJSON(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.ResetPassword(input)
	ctx.JSON(httpCode, codeMessage)
}

// Duplicate
// @Summary 檢查使用者是否重複
// @description 檢查使用者是否重複
// @Tags user
// @version 1.0
// @Accept json
// @produce json
// @param * body users.Filter true "檢查使用者是否重複"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /users/check-duplicate [post]
func (c *control) Duplicate(ctx *gin.Context) {
	input := &userModel.Field{}
	if err := ctx.ShouldBindJSON(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.Duplicate(input)
	ctx.JSON(httpCode, codeMessage)
}

// EnableAuthenticator
// @Summary 啟用驗證器
// @description 啟用驗證器
// @Tags user
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string  true "JWE Token"
// @param * body users.EnableAuthenticator true "啟用驗證器"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /users/authenticator/current-user [post]
func (c *control) EnableAuthenticator(ctx *gin.Context) {
	input := &userModel.EnableAuthenticator{}
	input.ID = ctx.MustGet("user_id").(string)
	if err := ctx.ShouldBindJSON(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.EnableAuthenticator(input)
	ctx.JSON(httpCode, codeMessage)
}

// ChangeEmail
// @Summary 更換電子郵件
// @description 更換電子郵件
// @Tags user
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string  true "JWE Token"
// @param * body users.ChangeEmail true "更換電子郵件"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /users/change-email/current-user [post]
func (c *control) ChangeEmail(ctx *gin.Context) {
	input := &userModel.ChangeEmail{}
	input.ID = ctx.MustGet("user_id").(string)
	if err := ctx.ShouldBindJSON(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.ChangeEmail(input)
	ctx.JSON(httpCode, codeMessage)
}

// VerifyEmail
// @Summary 驗證電子郵件
// @description 驗證電子郵件
// @Tags user
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string  true "JWE Token"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /users/verify-email/current-user [post]
func (c *control) VerifyEmail(ctx *gin.Context) {
	trx := ctx.MustGet("db_trx").(*gorm.DB)
	input := &userModel.VerifyEmail{}
	input.ID = ctx.MustGet("user_id").(string)
	input.Email = util.PointerString(ctx.MustGet("email").(string))
	input.UpdatedBy = util.PointerString(ctx.MustGet("user_id").(string))
	if err := ctx.ShouldBindQuery(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.VerifyEmail(trx, input)
	ctx.JSON(httpCode, codeMessage)
}
