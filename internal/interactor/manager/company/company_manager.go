package company

import (
	"encoding/json"
	"errors"
	"hta/internal/interactor/pkg/util"

	"gorm.io/gorm"

	companyModel "hta/internal/interactor/models/companies"
	companyService "hta/internal/interactor/service/company"

	"hta/internal/interactor/pkg/util/code"
	"hta/internal/interactor/pkg/util/log"
)

type Manager interface {
	Create(trx *gorm.DB, input *companyModel.Create) (int, any)
	GetByList(input *companyModel.Fields) (int, any)
	GetByListNoPagination(input *companyModel.Field) (int, any)
	GetBySingle(input *companyModel.Field) (int, any)
	Delete(input *companyModel.Field) (int, any)
	Update(input *companyModel.Update) (int, any)
}

type manager struct {
	CompanyService companyService.Service
}

func Init(db *gorm.DB) Manager {
	return &manager{
		CompanyService: companyService.Init(db),
	}
}

func (m *manager) Create(trx *gorm.DB, input *companyModel.Create) (int, any) {
	defer trx.Rollback()

	companyBase, err := m.CompanyService.WithTrx(trx).Create(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, companyBase.ID)
}

func (m *manager) GetByList(input *companyModel.Fields) (int, any) {
	output := &companyModel.List{}
	output.Limit = input.Limit
	output.Page = input.Page
	quantity, companyBase, err := m.CompanyService.GetByList(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	output.Total.Total = quantity
	output.Pages = util.Pagination(quantity, output.Limit)
	companyByte, err := json.Marshal(companyBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = json.Unmarshal(companyByte, &output.Companies)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	for i, company := range output.Companies {
		company.CreatedBy = *companyBase[i].CreatedByUsers.Name
		company.UpdatedBy = *companyBase[i].UpdatedByUsers.Name
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetByListNoPagination(input *companyModel.Field) (int, any) {
	output := &companyModel.List{}
	quantity, companyBase, err := m.CompanyService.GetByListNoPagination(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	output.Total.Total = quantity
	companyByte, err := json.Marshal(companyBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = json.Unmarshal(companyByte, &output.Companies)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	for i, company := range output.Companies {
		company.CreatedBy = *companyBase[i].CreatedByUsers.Name
		company.UpdatedBy = *companyBase[i].UpdatedByUsers.Name
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetBySingle(input *companyModel.Field) (int, any) {
	companyBase, err := m.CompanyService.GetBySingle(input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output := &companyModel.Single{}
	companyByte, _ := json.Marshal(companyBase)
	err = json.Unmarshal(companyByte, &output)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output.CreatedBy = *companyBase.CreatedByUsers.Name
	output.UpdatedBy = *companyBase.UpdatedByUsers.Name

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) Delete(input *companyModel.Field) (int, any) {
	_, err := m.CompanyService.GetBySingle(&companyModel.Field{
		ID: input.ID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.CompanyService.Delete(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, "Delete ok!")
}

func (m *manager) Update(input *companyModel.Update) (int, any) {
	companyBase, err := m.CompanyService.GetBySingle(&companyModel.Field{
		ID: input.ID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.CompanyService.Update(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, companyBase.ID)
}
