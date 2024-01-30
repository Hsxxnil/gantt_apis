package s3_file

import (
	"net/http"

	"hta/internal/interactor/manager/s3_file"
	s3FileModel "hta/internal/interactor/models/s3_files"
	"hta/internal/interactor/pkg/util/code"
	"hta/internal/interactor/pkg/util/log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Control interface {
	Create(ctx *gin.Context)
	Delete(ctx *gin.Context)
}

type control struct {
	Manager s3_file.Manager
}

func Init(db *gorm.DB) Control {
	return &control{
		Manager: s3_file.Init(db),
	}
}

// Create
// @Summary 上傳檔案
// @description 上傳檔案
// @Tags s3_file
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param * body s3_files.Create true "上傳檔案"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /files [post]
func (c *control) Create(ctx *gin.Context) {
	trx := ctx.MustGet("db_trx").(*gorm.DB)
	var input []*s3FileModel.Create
	if err := ctx.ShouldBindJSON(&input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	for _, create := range input {
		create.CreatedBy = ctx.MustGet("user_id").(string)
	}

	httpCode, codeMessage := c.Manager.Create(trx, input)
	ctx.JSON(httpCode, codeMessage)
}

// Delete
// @Summary 刪除單一檔案
// @description 刪除單一檔案
// @Tags s3_file
// @version 1.0
// @Accept json
// @produce json
// @param Authorization header string true "JWE Token"
// @param id path string true "檔案UUID"
// @success 200 object code.SuccessfulMessage{body=string} "成功後返回的值"
// @failure 415 object code.ErrorMessage{detailed=string} "必要欄位帶入錯誤"
// @failure 500 object code.ErrorMessage{detailed=string} "伺服器非預期錯誤"
// @Router /files/{id} [delete]
func (c *control) Delete(ctx *gin.Context) {
	trx := ctx.MustGet("db_trx").(*gorm.DB)
	id := ctx.Param("id")
	input := &s3FileModel.Field{}
	input.ID = id
	if err := ctx.ShouldBindQuery(input); err != nil {
		log.Error(err)
		ctx.JSON(http.StatusUnsupportedMediaType, code.GetCodeMessage(code.FormatError, err.Error()))
		return
	}

	httpCode, codeMessage := c.Manager.Delete(trx, input)
	ctx.JSON(httpCode, codeMessage)
}
