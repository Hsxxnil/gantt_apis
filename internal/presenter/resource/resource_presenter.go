package resource

import (
	"bytes"
	"encoding/csv"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"hta/internal/interactor/pkg/util"
	"hta/internal/interactor/pkg/util/hash"
	"net/http"
	"strconv"

	constant "hta/internal/interactor/constants"

	"hta/internal/interactor/manager/resource"
	resourceModel "hta/internal/interactor/models/resources"
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
	Import(ctx *gin.Context)
}

type control struct {
	Manager resource.Manager
}

func Init(db *gorm.DB) Control {
	return &control{
		Manager: resource.Init(db),
	}
}

// Create
// @Summary 新增資源
// @description 新增資源
// @Tags resource
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param * body resources.Create true "新增資源"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /resources [post]
func (c *control) Create(ctx *gin.Context) {
	trx := ctx.MustGet("db_trx").(*gorm.DB)
	input := &resourceModel.Create{}
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
// @Summary 取得全部資源
// @description 取得全部資源
// @Tags resource
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param page query int true "目前頁數,請從1開始帶入"
// @param limit query int true "一次回傳比數,請從1開始帶入,最高上限20"
// @param * body resources.Filter false "搜尋"
// @success 200 object code.SuccessfulMessage{body=resources.List} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /resources/list [post]
func (c *control) GetByList(ctx *gin.Context) {
	input := &resourceModel.Fields{}
	limit := ctx.Query("limit")
	page := ctx.Query("page")
	input.Limit, _ = strconv.ParseInt(limit, 10, 64)
	input.Page, _ = strconv.ParseInt(page, 10, 64)
	input.Role = util.PointerString(ctx.MustGet("role").(string))
	input.CreatedBy = util.PointerString(ctx.MustGet("user_id").(string))
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
// @Summary 取得全部資源(不用page&limit)
// @description 取得全部資源
// @Tags resource
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @success 200 object code.SuccessfulMessage{body=resources.List} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /resources/no-pagination [get]
func (c *control) GetByListNoPagination(ctx *gin.Context) {
	input := &resourceModel.Field{}
	if err := ctx.ShouldBindQuery(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.GetByListNoPagination(input)
	ctx.JSON(httpCode, codeMessage)
}

// GetBySingle
// @Summary 取得單一資源
// @description 取得單一資源
// @Tags resource
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param resource-uuid path string true "資源UUID"
// @success 200 object code.SuccessfulMessage{body=resources.Single} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /resources/{resource-uuid} [get]
func (c *control) GetBySingle(ctx *gin.Context) {
	resourceUUID := ctx.Param("resourceUUID")
	input := &resourceModel.Field{}
	input.ResourceUUID = resourceUUID
	if err := ctx.ShouldBindQuery(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.GetBySingle(input)
	ctx.JSON(httpCode, codeMessage)
}

// Delete
// @Summary 刪除單一資源
// @description 刪除單一資源
// @Tags resource
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param resource-uuid path string true "資源UUID"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /resources/{resource-uuid} [delete]
func (c *control) Delete(ctx *gin.Context) {
	resourceUUID := ctx.Param("resourceUUID")
	input := &resourceModel.Update{}
	input.ResourceUUID = resourceUUID
	input.Role = util.PointerString(ctx.MustGet("role").(string))
	if err := ctx.ShouldBindQuery(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.Delete(input)
	ctx.JSON(httpCode, codeMessage)
}

// Update
// @Summary 更新單一資源
// @description 更新單一資源
// @Tags resource
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param resource-uuid path string true "資源UUID"
// @param * body resources.Update true "更新資源"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /resources/{resource-uuid} [patch]
func (c *control) Update(ctx *gin.Context) {
	resourceUUID := ctx.Param("resourceUUID")
	input := &resourceModel.Update{}
	input.ResourceUUID = resourceUUID
	input.UpdatedBy = util.PointerString(ctx.MustGet("user_id").(string))
	input.Role = util.PointerString(ctx.MustGet("role").(string))
	if err := ctx.ShouldBindJSON(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.Update(input)
	ctx.JSON(httpCode, codeMessage)
}

// Import
// @Summary 匯入資源
// @description 匯入資源
// @Tags resource
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param * body resources.Import true "匯入資源"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /resources/import [post]
func (c *control) Import(ctx *gin.Context) {
	input := &resourceModel.Import{}
	trx := ctx.MustGet("db_trx").(*gorm.DB)
	if err := ctx.ShouldBindJSON(&input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	input.CreatedBy = ctx.MustGet("user_id").(string)
	inputByte := hash.Base64StdDecode(input.Base64)
	readerFile := csv.NewReader(transform.NewReader(bytes.NewBuffer(inputByte), unicode.UTF8.NewDecoder()))
	input.CSVFile = readerFile

	httpCode, codeMessage := c.Manager.Import(trx, input)
	ctx.JSON(httpCode, codeMessage)
}
