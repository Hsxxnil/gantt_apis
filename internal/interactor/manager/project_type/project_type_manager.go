package project_type

import (
	"errors"
	"github.com/bytedance/sonic"
	"hta/internal/interactor/pkg/util"

	"gorm.io/gorm"

	projectTypeModel "hta/internal/interactor/models/project_types"
	projectTypeService "hta/internal/interactor/service/project_type"

	"hta/internal/interactor/pkg/util/code"
	"hta/internal/interactor/pkg/util/log"
)

type Manager interface {
	Create(trx *gorm.DB, input *projectTypeModel.Create) (int, any)
	GetByList(input *projectTypeModel.Fields) (int, any)
	GetByListNoPagination(input *projectTypeModel.Field) (int, any)
	GetBySingle(input *projectTypeModel.Field) (int, any)
	Delete(input *projectTypeModel.Field) (int, any)
	Update(input *projectTypeModel.Update) (int, any)
}

type manager struct {
	ProjectTypeService projectTypeService.Service
}

func Init(db *gorm.DB) Manager {
	return &manager{
		ProjectTypeService: projectTypeService.Init(db),
	}
}

func (m *manager) Create(trx *gorm.DB, input *projectTypeModel.Create) (int, any) {
	defer trx.Rollback()

	projectTypeBase, err := m.ProjectTypeService.WithTrx(trx).Create(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, projectTypeBase.ID)
}

func (m *manager) GetByList(input *projectTypeModel.Fields) (int, any) {
	output := &projectTypeModel.List{}
	output.Limit = input.Limit
	output.Page = input.Page
	quantity, projectTypeBase, err := m.ProjectTypeService.GetByList(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	output.Total.Total = quantity
	output.Pages = util.Pagination(quantity, output.Limit)
	projectTypeByte, err := sonic.Marshal(projectTypeBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(projectTypeByte, &output.ProjectTypes)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetByListNoPagination(input *projectTypeModel.Field) (int, any) {
	output := &projectTypeModel.List{}
	quantity, projectTypeBase, err := m.ProjectTypeService.GetByListNoPagination(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	output.Total.Total = quantity
	projectTypeByte, err := sonic.Marshal(projectTypeBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(projectTypeByte, &output.ProjectTypes)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetBySingle(input *projectTypeModel.Field) (int, any) {
	projectTypeBase, err := m.ProjectTypeService.GetBySingle(input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output := &projectTypeModel.Single{}
	projectTypeByte, _ := sonic.Marshal(projectTypeBase)
	err = sonic.Unmarshal(projectTypeByte, &output)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) Delete(input *projectTypeModel.Field) (int, any) {
	_, err := m.ProjectTypeService.GetBySingle(&projectTypeModel.Field{
		ID: input.ID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.ProjectTypeService.Delete(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, "Delete ok!")
}

func (m *manager) Update(input *projectTypeModel.Update) (int, any) {
	projectTypeBase, err := m.ProjectTypeService.GetBySingle(&projectTypeModel.Field{
		ID: input.ID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.ProjectTypeService.Update(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, projectTypeBase.ID)
}
