package policy

import (
	"gantt/internal/interactor/manager/policy"
	policyModel "gantt/internal/interactor/models/policies"
	"net/http"

	"gantt/internal/interactor/pkg/util/code"
	"gantt/internal/interactor/pkg/util/log"

	"github.com/gin-gonic/gin"
	_ "gorm.io/gorm"
)

type Control interface {
	Create(ctx *gin.Context)
	GetByList(ctx *gin.Context)
	Delete(ctx *gin.Context)
}

type control struct {
	Manager policy.Manager
}

func Init() Control {
	return &control{
		Manager: policy.Init(),
	}
}

// Create
// @Summary 新增策略
// @description 新增策略
// @Tags policies
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string  true "JWE Token"
// @param * body []policies.PolicyRule true "新增策略"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /policies [post]
func (c *control) Create(ctx *gin.Context) {
	var input []*policyModel.PolicyRule
	if err := ctx.ShouldBindJSON(&input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.Create(input)
	ctx.JSON(httpCode, codeMessage)
}

// GetByList
// @Summary 取得策略
// @description 取得策略
// @Tags policies
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string  true "JWE Token"
// @success 200 object code.SuccessfulMessage{body=[]policies.Single} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /policies [get]
func (c *control) GetByList(ctx *gin.Context) {
	httpCode, codeMessage := c.Manager.GetByList()
	ctx.JSON(httpCode, codeMessage)
}

// Delete
// @Summary 刪除策略
// @description 刪除策略
// @Tags policies
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string  true "JWE Token"
// @param * body []policies.PolicyRule true "刪除策略"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /policies [delete]
func (c *control) Delete(ctx *gin.Context) {
	var input []*policyModel.PolicyRule
	if err := ctx.ShouldBindJSON(&input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.Delete(input)
	ctx.JSON(httpCode, codeMessage)
}
