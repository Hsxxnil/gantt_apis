package company

import (
	"hta/internal/interactor/pkg/util"
	"net/http"

	constant "hta/internal/interactor/constants"

	"hta/internal/interactor/manager/company"
	companyModel "hta/internal/interactor/models/companies"
	"hta/internal/interactor/pkg/util/code"
	"hta/internal/interactor/pkg/util/log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Control interface {
	Create(ctx *gin.Context)
	GetByList(ctx *gin.Context)
	GetByListNoPagination(ctx *gin.Context)
	GetBySingle(ctx *gin.Context)
	Delete(ctx *gin.Context)
	Update(ctx *gin.Context)
}

type control struct {
	Manager company.Manager
}

func Init(db *gorm.DB) Control {
	return &control{
		Manager: company.Init(db),
	}
}

// Create
// @Summary 新增公司
// @description 新增公司
// @Tags company
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param * body companies.Create true "新增公司"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /companies [post]
func (c *control) Create(ctx *gin.Context) {
	trx := ctx.MustGet("db_trx").(*gorm.DB)
	input := &companyModel.Create{}
	input.CreatedBy = ctx.MustGet("user_id").(string)
	if err := ctx.ShouldBindJSON(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))

		return
	}

	httpCode, codeMessage := c.Manager.Create(trx, input)
	ctx.JSON(httpCode, codeMessage)
}

// GetByList
// @Summary 取得全部公司
// @description 取得全部公司
// @Tags company
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param page query int true "目前頁數,請從1開始帶入"
// @param limit query int true "一次回傳比數,請從1開始帶入,最高上限20"
// @success 200 object code.SuccessfulMessage{body=companies.List} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /companies [get]
func (c *control) GetByList(ctx *gin.Context) {
	input := &companyModel.Fields{}

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
// @Summary 取得全部公司(不用page&limit)
// @description 取得全部公司
// @Tags company
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @success 200 object code.SuccessfulMessage{body=tasks.List} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /companies/no-pagination [get]
func (c *control) GetByListNoPagination(ctx *gin.Context) {
	input := &companyModel.Field{}
	if err := ctx.ShouldBindQuery(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))

		return
	}

	httpCode, codeMessage := c.Manager.GetByListNoPagination(input)
	ctx.JSON(httpCode, codeMessage)
}

// GetBySingle
// @Summary 取得單一公司
// @description 取得單一公司
// @Tags company
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param id path string true "公司UUID"
// @success 200 object code.SuccessfulMessage{body=companies.Single} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /companies/{id} [get]
func (c *control) GetBySingle(ctx *gin.Context) {
	id := ctx.Param("id")
	input := &companyModel.Field{}
	input.ID = id
	if err := ctx.ShouldBindQuery(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))

		return
	}

	httpCode, codeMessage := c.Manager.GetBySingle(input)
	ctx.JSON(httpCode, codeMessage)
}

// Delete
// @Summary 刪除單一公司
// @description 刪除單一公司
// @Tags company
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param id path string true "公司UUID"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /companies/{id} [delete]
func (c *control) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	input := &companyModel.Field{}
	input.ID = id
	if err := ctx.ShouldBindQuery(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))

		return
	}

	httpCode, codeMessage := c.Manager.Delete(input)
	ctx.JSON(httpCode, codeMessage)
}

// Update
// @Summary 更新單一公司
// @description 更新單一公司
// @Tags company
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param id path string true "公司UUID"
// @param * body companies.Update true "更新公司"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /companies/{id} [patch]
func (c *control) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	input := &companyModel.Update{}
	input.ID = id
	input.UpdatedBy = util.PointerString(ctx.MustGet("user_id").(string))
	if err := ctx.ShouldBindJSON(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))

		return
	}

	httpCode, codeMessage := c.Manager.Update(input)
	ctx.JSON(httpCode, codeMessage)
}
