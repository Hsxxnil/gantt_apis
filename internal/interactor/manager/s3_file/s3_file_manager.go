package s3_file

import (
	"errors"
	"gorm.io/gorm"
	"hta/internal/interactor/pkg/util"
	"path/filepath"

	s3FileModel "hta/internal/interactor/models/s3_files"
	s3FileService "hta/internal/interactor/service/s3_file"

	"hta/internal/interactor/pkg/util/code"
	"hta/internal/interactor/pkg/util/log"
)

type Manager interface {
	Create(trx *gorm.DB, input *s3FileModel.Create) (int, any)
	Delete(trx *gorm.DB, input *s3FileModel.Field) (int, any)
}

type manager struct {
	S3FileService s3FileService.Service
}

func Init(db *gorm.DB) Manager {
	return &manager{
		S3FileService: s3FileService.Init(db),
	}
}

func (m *manager) Create(trx *gorm.DB, input *s3FileModel.Create) (int, any) {
	defer trx.Rollback()
	// confirm s3 bucket name
	s3BucketName := "myhta"
	input.FileExtension = filepath.Ext(input.FileName)
	filePath := "files/" + input.SourceUUID + "/" + input.FileName

	// upload file to s3
	url, err := util.UploadToS3(input.Base64, filePath, s3BucketName)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	if url == "" {
		log.Error("Upload to s3 failed.")
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, "Upload to s3 failed.")
	}

	input.FileUrl = url
	s3FileBase, err := m.S3FileService.WithTrx(trx).Create(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, s3FileBase.ID)
}

func (m *manager) Delete(trx *gorm.DB, input *s3FileModel.Field) (int, any) {
	defer trx.Rollback()

	_, err := m.S3FileService.GetBySingle(&s3FileModel.Field{
		ID: input.ID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.S3FileService.Delete(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, "Delete ok!")
}
