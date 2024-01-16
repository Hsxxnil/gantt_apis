package work_day

import (
	"hta/internal/interactor/pkg/util"
	"net/http"

	constant "hta/internal/interactor/constants"

	"hta/internal/interactor/manager/work_day"
	workDayModel "hta/internal/interactor/models/work_days"
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
	Manager work_day.Manager
}

func Init(db *gorm.DB) Control {
	return &control{
		Manager: work_day.Init(db),
	}
}

// Create
// @Summary 新增工作時間
// @description 新增工作時間
// @Tags work_day
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param * body work_days.Create true "新增工作時間"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /work-days [post]
func (c *control) Create(ctx *gin.Context) {
	trx := ctx.MustGet("db_trx").(*gorm.DB)
	input := &workDayModel.Create{}
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
// @Summary 取得全部工作時間
// @description 取得全部工作時間
// @Tags work_day
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param page query int true "目前頁數,請從1開始帶入"
// @param limit query int true "一次回傳比數,請從1開始帶入,最高上限20"
// @success 200 object code.SuccessfulMessage{body=work_days.List} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /work-days [get]
func (c *control) GetByList(ctx *gin.Context) {
	input := &workDayModel.Fields{}

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
// @Summary 取得全部工作時間(不用page&limit)
// @description 取得全部工作時間
// @Tags work_day
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @success 200 object code.SuccessfulMessage{body=work_days.List} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /work-days/no-pagination [get]
func (c *control) GetByListNoPagination(ctx *gin.Context) {
	input := &workDayModel.Field{}
	if err := ctx.ShouldBindQuery(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.GetByListNoPagination(input)
	ctx.JSON(httpCode, codeMessage)
}

// GetBySingle
// @Summary 取得單一工作時間
// @description 取得單一工作時間
// @Tags work_day
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param id path string true "工作時間UUID"
// @success 200 object code.SuccessfulMessage{body=work_days.Single} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /work-days/{id} [get]
func (c *control) GetBySingle(ctx *gin.Context) {
	id := ctx.Param("id")
	input := &workDayModel.Field{}
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
// @Summary 刪除單一工作時間
// @description 刪除單一工作時間
// @Tags work_day
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param id path string true "工作時間UUID"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /work-days/{id} [delete]
func (c *control) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	input := &workDayModel.Field{}
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
// @Summary 更新單一工作時間
// @description 更新單一工作時間
// @Tags work_day
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param id path string true "工作時間UUID"
// @param * body work_days.Update true "更新工作時間"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /work-days/{id} [patch]
func (c *control) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	input := &workDayModel.Update{}
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
