package department

import (
	"errors"
	"github.com/bytedance/sonic"
	"hta/internal/interactor/pkg/util"

	"gorm.io/gorm"

	departmentModel "hta/internal/interactor/models/departments"
	departmentService "hta/internal/interactor/service/department"

	"hta/internal/interactor/pkg/util/code"
	"hta/internal/interactor/pkg/util/log"
)

type Manager interface {
	Create(trx *gorm.DB, input *departmentModel.Create) (int, any)
	GetByList(input *departmentModel.Fields) (int, any)
	GetByListNoPagination(input *departmentModel.Field) (int, any)
	GetBySingle(input *departmentModel.Field) (int, any)
	Delete(input *departmentModel.Field) (int, any)
	Update(input *departmentModel.Update) (int, any)
}

type manager struct {
	DepartmentService departmentService.Service
}

func Init(db *gorm.DB) Manager {
	return &manager{
		DepartmentService: departmentService.Init(db),
	}
}

func (m *manager) Create(trx *gorm.DB, input *departmentModel.Create) (int, any) {
	defer trx.Rollback()

	departmentBase, err := m.DepartmentService.WithTrx(trx).Create(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, departmentBase.ID)
}

func (m *manager) GetByList(input *departmentModel.Fields) (int, any) {
	output := &departmentModel.List{}
	output.Limit = input.Limit
	output.Page = input.Page
	quantity, departmentBase, err := m.DepartmentService.GetByList(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	output.Total.Total = quantity
	output.Pages = util.Pagination(quantity, output.Limit)
	departmentByte, err := sonic.Marshal(departmentBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(departmentByte, &output.Departments)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	for i, department := range output.Departments {
		department.CreatedBy = *departmentBase[i].CreatedByUsers.Name
		department.UpdatedBy = *departmentBase[i].UpdatedByUsers.Name
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetByListNoPagination(input *departmentModel.Field) (int, any) {
	output := &departmentModel.List{}
	quantity, departmentBase, err := m.DepartmentService.GetByListNoPagination(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	output.Total.Total = quantity
	departmentByte, err := sonic.Marshal(departmentBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(departmentByte, &output.Departments)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	for i, department := range output.Departments {
		department.CreatedBy = *departmentBase[i].CreatedByUsers.Name
		department.UpdatedBy = *departmentBase[i].UpdatedByUsers.Name
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetBySingle(input *departmentModel.Field) (int, any) {
	departmentBase, err := m.DepartmentService.GetBySingle(input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output := &departmentModel.Single{}
	departmentByte, _ := sonic.Marshal(departmentBase)
	err = sonic.Unmarshal(departmentByte, &output)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output.CreatedBy = *departmentBase.CreatedByUsers.Name
	output.UpdatedBy = *departmentBase.UpdatedByUsers.Name

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) Delete(input *departmentModel.Field) (int, any) {
	_, err := m.DepartmentService.GetBySingle(&departmentModel.Field{
		ID: input.ID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.DepartmentService.Delete(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, "Delete ok!")
}

func (m *manager) Update(input *departmentModel.Update) (int, any) {
	departmentBase, err := m.DepartmentService.GetBySingle(&departmentModel.Field{
		ID: input.ID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.DepartmentService.Update(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, departmentBase.ID)
}
