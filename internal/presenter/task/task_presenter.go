package task

import (
	"bytes"
	"encoding/csv"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"hta/internal/interactor/pkg/util"
	"hta/internal/interactor/pkg/util/hash"
	"net/http"

	"hta/internal/interactor/manager/task"
	taskModel "hta/internal/interactor/models/tasks"
	"hta/internal/interactor/pkg/util/code"
	"hta/internal/interactor/pkg/util/log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Control interface {
	Create(ctx *gin.Context)
	CreateAll(ctx *gin.Context)
	GetByProjectUUIDList(ctx *gin.Context)
	GetByListNoPaginationNoSub(ctx *gin.Context)
	GetBySingle(ctx *gin.Context)
	Delete(ctx *gin.Context)
	Update(ctx *gin.Context)
	UpdateAll(ctx *gin.Context)
	Import(ctx *gin.Context)
}

type control struct {
	Manager task.Manager
}

func Init(db *gorm.DB) Control {
	return &control{
		Manager: task.Init(db),
	}
}

// Create
// @Summary 新增單一任務
// @description 新增單一任務
// @Tags task
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param * body tasks.Create true "新增任務"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /tasks [post]
func (c *control) Create(ctx *gin.Context) {
	trx := ctx.MustGet("db_trx").(*gorm.DB)
	input := &taskModel.Create{}
	input.CreatedBy = ctx.MustGet("user_id").(string)
	input.Role = util.PointerString(ctx.MustGet("role").(string))
	input.ResUUID = util.PointerString(ctx.MustGet("resource_id").(string))
	if err := ctx.ShouldBindJSON(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.Create(trx, input)
	ctx.JSON(httpCode, codeMessage)
}

// CreateAll
// @Summary 新增全任務
// @description 新增全任務
// @Tags task
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param * body []tasks.Create true "新增任務"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /tasks/create-all [post]
func (c *control) CreateAll(ctx *gin.Context) {
	trx := ctx.MustGet("db_trx").(*gorm.DB)
	var input []*taskModel.Create
	if err := ctx.ShouldBindJSON(&input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	for _, create := range input {
		create.CreatedBy = ctx.MustGet("user_id").(string)
		create.Role = util.PointerString(ctx.MustGet("role").(string))
		create.ResUUID = util.PointerString(ctx.MustGet("resource_id").(string))
	}

	httpCode, codeMessage := c.Manager.CreateAll(trx, input)
	ctx.JSON(httpCode, codeMessage)
}

// GetByListNoPaginationNoSub
// @Summary 取得全部不算階層的任務(不用page&limit)
// @description 取得全部不算階層的任務(不用page&limit)
// @Tags task
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @success 200 object code.SuccessfulMessage{body=tasks.List} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /tasks/no-pagination/no-sub-filter [get]
func (c *control) GetByListNoPaginationNoSub(ctx *gin.Context) {
	input := &taskModel.Field{}
	if err := ctx.ShouldBindQuery(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.GetByListNoPaginationNoSub(input)
	ctx.JSON(httpCode, codeMessage)
}

// GetByProjectUUIDList
// @Summary 取得多個專案含任務(不用page&limit)
// @description 取得多個專案含任務(不用page&limit)
// @Tags task
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param * body tasks.ProjectIDs true "專案UUIDs"
// @param * body tasks.Filter false "搜尋"
// @success 200 object code.SuccessfulMessage{body=tasks.List} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /tasks/get-by-projects [post]
func (c *control) GetByProjectUUIDList(ctx *gin.Context) {
	input := &taskModel.ProjectIDs{}
	input.Role = util.PointerString(ctx.MustGet("role").(string))
	input.UserID = util.PointerString(ctx.MustGet("user_id").(string))
	input.ResUUID = util.PointerString(ctx.MustGet("resource_id").(string))
	if err := ctx.ShouldBindJSON(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.GetByProjectListNoPagination(input)
	ctx.JSON(httpCode, codeMessage)
}

// GetBySingle
// @Summary 取得單一任務
// @description 取得單一任務
// @Tags task
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param task-uuid path string true "任務UUID"
// @success 200 object code.SuccessfulMessage{body=tasks.Single} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /tasks/{task-uuid} [get]
func (c *control) GetBySingle(ctx *gin.Context) {
	taskUUID := ctx.Param("taskUUID")
	input := &taskModel.Field{}
	input.TaskUUID = taskUUID
	if err := ctx.ShouldBindQuery(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.GetBySingle(input)
	ctx.JSON(httpCode, codeMessage)
}

// Delete
// @Summary 刪除單一任務
// @description 刪除單一任務
// @Tags task
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param task-uuid path string true "任務UUID"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /tasks [delete]
func (c *control) Delete(ctx *gin.Context) {
	trx := ctx.MustGet("db_trx").(*gorm.DB)
	input := &taskModel.DeletedTaskUUIDs{}
	input.Role = util.PointerString(ctx.MustGet("role").(string))
	input.ResUUID = util.PointerString(ctx.MustGet("resource_id").(string))
	if err := ctx.ShouldBindJSON(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.Delete(trx, input)
	ctx.JSON(httpCode, codeMessage)
}

// Update
// @Summary 更新單一任務
// @description 更新單一任務
// @Tags task
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param task-uuid path string true "任務UUID"
// @param * body tasks.Update true "更新任務"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /tasks/{task-uuid} [patch]
func (c *control) Update(ctx *gin.Context) {
	trx := ctx.MustGet("db_trx").(*gorm.DB)
	taskUUID := ctx.Param("taskUUID")
	input := &taskModel.Update{}
	input.TaskUUID = taskUUID
	input.UpdatedBy = util.PointerString(ctx.MustGet("user_id").(string))
	input.Role = util.PointerString(ctx.MustGet("role").(string))
	input.ResUUID = util.PointerString(ctx.MustGet("resource_id").(string))
	if err := ctx.ShouldBindJSON(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.Update(trx, input)
	ctx.JSON(httpCode, codeMessage)
}

// UpdateAll
// @Summary 更新全任務
// @description 更新全任務
// @Tags task
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param * body []tasks.Update true "更新任務"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /tasks/update-all [patch]
func (c *control) UpdateAll(ctx *gin.Context) {
	trx := ctx.MustGet("db_trx").(*gorm.DB)
	var input []*taskModel.Update
	if err := ctx.ShouldBindJSON(&input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	for _, update := range input {
		update.UpdatedBy = util.PointerString(ctx.MustGet("user_id").(string))
		update.Role = util.PointerString(ctx.MustGet("role").(string))
		update.ResUUID = util.PointerString(ctx.MustGet("resource_id").(string))
	}

	httpCode, codeMessage := c.Manager.UpdateAll(trx, input)
	ctx.JSON(httpCode, codeMessage)
}

// Import
// @Summary 匯入專案
// @description 匯入專案
// @Tags task
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param * body tasks.Import true "匯入專案"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /tasks/import [post]
func (c *control) Import(ctx *gin.Context) {
	input := &taskModel.Import{}
	trx := ctx.MustGet("db_trx").(*gorm.DB)
	input.CreatedBy = ctx.MustGet("user_id").(string)
	input.Role = util.PointerString(ctx.MustGet("role").(string))
	input.ResUUID = util.PointerString(ctx.MustGet("resource_id").(string))
	if err := ctx.ShouldBindJSON(&input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	inputByte := hash.Base64StdDecode(input.Base64)
	readerFile := csv.NewReader(transform.NewReader(bytes.NewBuffer(inputByte), unicode.UTF8.NewDecoder()))
	input.CSVFile = readerFile

	httpCode, codeMessage := c.Manager.Import(trx, input)
	ctx.JSON(httpCode, codeMessage)
}
