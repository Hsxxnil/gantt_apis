package project_resource

import (
	"net/http"
	"strconv"

	constant "hta/internal/interactor/constants"

	"hta/internal/interactor/manager/project_resource"
	projectResourceModel "hta/internal/interactor/models/project_resources"
	"hta/internal/interactor/pkg/util/code"
	"hta/internal/interactor/pkg/util/log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Control interface {
	GetByList(ctx *gin.Context)
	GetByProjectList(ctx *gin.Context)
	GetBySingle(ctx *gin.Context)
}

type control struct {
	Manager project_resource.Manager
}

func Init(db *gorm.DB) Control {
	return &control{
		Manager: project_resource.Init(db),
	}
}

// GetByList
// @Summary 取得全部專案資源
// @description 取得全部專案資源
// @Tags project_resource
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param page query int true "目前頁數,請從1開始帶入"
// @param limit query int true "一次回傳比數,請從1開始帶入,最高上限20"
// @param * body project_resources.Filter false "搜尋"
// @success 200 object code.SuccessfulMessage{body=project_resources.List} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /project-resources [post]
func (c *control) GetByList(ctx *gin.Context) {
	input := &projectResourceModel.Fields{}
	limit := ctx.Query("limit")
	page := ctx.Query("page")
	input.Limit, _ = strconv.ParseInt(limit, 10, 64)
	input.Page, _ = strconv.ParseInt(page, 10, 64)

	if err := ctx.ShouldBindJSON(input); err != nil {
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

// GetByProjectList
// @Summary 透過過專案ID取得專案資源(不用page&limit)
// @description 透過過專案ID取得專案資源
// @Tags project_resource
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param project-uuid path string true "專案UUID"
// @success 200 object code.SuccessfulMessage{body=tasks.List} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /project-resources/get-by-project [post]
func (c *control) GetByProjectList(ctx *gin.Context) {
	input := &projectResourceModel.ProjectIDs{}
	if err := ctx.ShouldBindJSON(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))

		return
	}

	httpCode, codeMessage := c.Manager.GetByProjectList(input)
	ctx.JSON(httpCode, codeMessage)
}

// GetBySingle
// @Summary 取得單一專案資源
// @description 取得單一專案資源
// @Tags project_resource
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param id path string true "專案資源UUID"
// @success 200 object code.SuccessfulMessage{body=project_resources.Single} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /project-resources/{id} [get]
func (c *control) GetBySingle(ctx *gin.Context) {
	id := ctx.Param("id")
	input := &projectResourceModel.Field{}
	input.ID = id
	if err := ctx.ShouldBindQuery(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))

		return
	}

	httpCode, codeMessage := c.Manager.GetBySingle(input)
	ctx.JSON(httpCode, codeMessage)
}
