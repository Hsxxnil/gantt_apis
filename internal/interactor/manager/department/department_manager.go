package department

import (
	"errors"
	"github.com/bytedance/sonic"
	affiliationModel "hta/internal/interactor/models/affiliations"
	"hta/internal/interactor/pkg/util"
	affiliationService "hta/internal/interactor/service/affiliation"

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
	Delete(trx *gorm.DB, input *departmentModel.Field) (int, any)
	Update(trx *gorm.DB, input *departmentModel.Update) (int, any)
}

type manager struct {
	DepartmentService  departmentService.Service
	AffiliationService affiliationService.Service
}

func Init(db *gorm.DB) Manager {
	return &manager{
		DepartmentService:  departmentService.Init(db),
		AffiliationService: affiliationService.Init(db),
	}
}

func (m *manager) Create(trx *gorm.DB, input *departmentModel.Create) (int, any) {
	defer trx.Rollback()

	departmentBase, err := m.DepartmentService.WithTrx(trx).Create(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// sync create affiliation
	if input.SupervisorID != nil {
		_, err = m.AffiliationService.WithTrx(trx).Create(&affiliationModel.Create{
			UserID:       *input.SupervisorID,
			DeptID:       *departmentBase.ID,
			IsSupervisor: true,
			CreatedBy:    input.CreatedBy,
		})
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
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
		for j, affiliation := range department.Affiliations {
			affiliation.Name = *departmentBase[i].Affiliations[j].Users.Name
		}
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetByListNoPagination(input *departmentModel.Field) (int, any) {
	output := &departmentModel.List{}
	departmentBase, err := m.DepartmentService.GetByListNoPagination(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
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
	for i, affiliation := range output.Affiliations {
		affiliation.Name = *departmentBase.Affiliations[i].Users.Name
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) Delete(trx *gorm.DB, input *departmentModel.Field) (int, any) {
	defer trx.Rollback()

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

	err = m.DepartmentService.WithTrx(trx).Delete(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// sync delete affiliation
	err = m.AffiliationService.WithTrx(trx).Delete(&affiliationModel.Field{
		DeptID: util.PointerString(input.ID),
	})
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, "Delete ok!")
}

func (m *manager) Update(trx *gorm.DB, input *departmentModel.Update) (int, any) {
	defer trx.Rollback()

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

	err = m.DepartmentService.WithTrx(trx).Update(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	if input.SupervisorID != nil {
		// check the original supervisor exist
		originalAffiliationBase, err := m.AffiliationService.GetBySingle(&affiliationModel.Field{
			DeptID:       util.PointerString(input.ID),
			IsSupervisor: util.PointerBool(true),
		})
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
		}

		if originalAffiliationBase != nil {
			// update the original supervisor's affiliation
			err = m.AffiliationService.WithTrx(trx).Update(&affiliationModel.Update{
				ID:           *originalAffiliationBase.ID,
				IsSupervisor: util.PointerBool(false),
				UpdatedBy:    input.UpdatedBy,
			})
			if err != nil {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
		}

		// sync update the new supervisor's affiliation
		err = m.AffiliationService.WithTrx(trx).Update(&affiliationModel.Update{
			DeptID:       util.PointerString(input.ID),
			UserID:       input.SupervisorID,
			IsSupervisor: util.PointerBool(true),
			UpdatedBy:    input.UpdatedBy,
		})
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, departmentBase.ID)
}
