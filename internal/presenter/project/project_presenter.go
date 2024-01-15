package project

import (
	"hta/internal/interactor/pkg/util"
	"net/http"
	"strconv"

	constant "hta/internal/interactor/constants"

	"hta/internal/interactor/manager/project"
	projectModel "hta/internal/interactor/models/projects"
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
	Manager project.Manager
}

func Init(db *gorm.DB) Control {
	return &control{
		Manager: project.Init(db),
	}
}

// Create
// @Summary 新增專案
// @description 新增專案
// @Tags project
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param * body projects.Create true "新增專案"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /projects [post]
func (c *control) Create(ctx *gin.Context) {
	trx := ctx.MustGet("db_trx").(*gorm.DB)
	input := &projectModel.Create{}
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
// @Summary 取得全部專案
// @description 取得全部專案
// @Tags project
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param page query int true "目前頁數,請從1開始帶入"
// @param limit query int true "一次回傳比數,請從1開始帶入,最高上限20"
// @param * body projects.Filter false "搜尋"
// @success 200 object code.SuccessfulMessage{body=projects.List} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /projects/list [post]
func (c *control) GetByList(ctx *gin.Context) {
	input := &projectModel.Fields{}
	limit := ctx.Query("limit")
	page := ctx.Query("page")
	input.Limit, _ = strconv.ParseInt(limit, 10, 64)
	input.Page, _ = strconv.ParseInt(page, 10, 64)
	if ctx.MustGet("role").(string) != "admin" {
		input.CreatedBy = util.PointerString(ctx.MustGet("user_id").(string))
		input.ResourceUUID = util.PointerString(ctx.MustGet("resource_id").(string))
	}
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

// GetByListNoPagination
// @Summary 取得全部專案(不用page&limit)
// @description 取得全部專案
// @Tags project
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @success 200 object code.SuccessfulMessage{body=projects.List} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /projects/no-pagination [get]
func (c *control) GetByListNoPagination(ctx *gin.Context) {
	input := &projectModel.Field{}
	if ctx.MustGet("role").(string) != "admin" {
		input.CreatedBy = util.PointerString(ctx.MustGet("user_id").(string))
		input.ResourceUUID = util.PointerString(ctx.MustGet("resource_id").(string))
	}
	if err := ctx.ShouldBindQuery(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))

		return
	}

	httpCode, codeMessage := c.Manager.GetByListNoPagination(input)
	ctx.JSON(httpCode, codeMessage)
}

// GetBySingle
// @Summary 取得單一專案
// @description 取得單一專案
// @Tags project
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param project-uuid path string true "專案UUID"
// @success 200 object code.SuccessfulMessage{body=projects.Single} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /projects/{project-uuid} [get]
func (c *control) GetBySingle(ctx *gin.Context) {
	projectID := ctx.Param("projectID")
	input := &projectModel.Field{}
	input.ProjectUUID = projectID
	if ctx.MustGet("role").(string) == "user" {
		input.CreatedBy = util.PointerString(ctx.MustGet("user_id").(string))
	}
	if err := ctx.ShouldBindQuery(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))

		return
	}

	httpCode, codeMessage := c.Manager.GetBySingle(input)
	ctx.JSON(httpCode, codeMessage)
}

// Delete
// @Summary 刪除單一專案
// @description 刪除單一專案
// @Tags project
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param project-uuid path string true "專案UUID"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /projects/{project-uuid} [delete]
func (c *control) Delete(ctx *gin.Context) {
	trx := ctx.MustGet("db_trx").(*gorm.DB)
	projectID := ctx.Param("projectID")
	input := &projectModel.Update{}
	input.ProjectUUID = projectID
	input.UpdatedBy = util.PointerString(ctx.MustGet("user_id").(string))
	if ctx.MustGet("role").(string) != "admin" {
		input.ResourceUUID = util.PointerString(ctx.MustGet("resource_id").(string))
	}
	if err := ctx.ShouldBindQuery(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))

		return
	}

	httpCode, codeMessage := c.Manager.Delete(trx, input)
	ctx.JSON(httpCode, codeMessage)
}

// Update
// @Summary 更新單一專案
// @description 更新單一專案
// @Tags project
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param project-uuid path string true "專案UUID"
// @param * body projects.Update true "更新專案"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /projects/{project-uuid} [patch]
func (c *control) Update(ctx *gin.Context) {
	trx := ctx.MustGet("db_trx").(*gorm.DB)
	projectID := ctx.Param("projectID")
	input := &projectModel.Update{}
	input.ProjectUUID = projectID
	input.UpdatedBy = util.PointerString(ctx.MustGet("user_id").(string))
	if ctx.MustGet("role").(string) != "admin" {
		input.ResourceUUID = util.PointerString(ctx.MustGet("resource_id").(string))
	}
	if err := ctx.ShouldBindJSON(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))

		return
	}

	httpCode, codeMessage := c.Manager.Update(trx, input)
	ctx.JSON(httpCode, codeMessage)
}
