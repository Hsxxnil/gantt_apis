package organization

import (
	"encoding/json"
	"errors"
	"hta/internal/interactor/pkg/util"

	"gorm.io/gorm"

	organizationModel "hta/internal/interactor/models/organizations"
	organizationService "hta/internal/interactor/service/organization"

	"hta/internal/interactor/pkg/util/code"
	"hta/internal/interactor/pkg/util/log"
)

type Manager interface {
	Create(trx *gorm.DB, input *organizationModel.Create) (int, any)
	GetByList(input *organizationModel.Fields) (int, any)
	GetByListNoPagination(input *organizationModel.Field) (int, any)
	GetBySingle(input *organizationModel.Field) (int, any)
	Delete(input *organizationModel.Field) (int, any)
	Update(input *organizationModel.Update) (int, any)
}

type manager struct {
	OrganizationService organizationService.Service
}

func Init(db *gorm.DB) Manager {
	return &manager{
		OrganizationService: organizationService.Init(db),
	}
}

func (m *manager) Create(trx *gorm.DB, input *organizationModel.Create) (int, any) {
	defer trx.Rollback()

	organizationBase, err := m.OrganizationService.WithTrx(trx).Create(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, organizationBase.ID)
}

func (m *manager) GetByList(input *organizationModel.Fields) (int, any) {
	output := &organizationModel.List{}
	output.Limit = input.Limit
	output.Page = input.Page
	quantity, organizationBase, err := m.OrganizationService.GetByList(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	output.Total.Total = quantity
	output.Pages = util.Pagination(quantity, output.Limit)
	organizationByte, err := json.Marshal(organizationBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = json.Unmarshal(organizationByte, &output.Organizations)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	for i, organization := range output.Organizations {
		organization.CreatedBy = *organizationBase[i].CreatedByUsers.Name
		organization.UpdatedBy = *organizationBase[i].UpdatedByUsers.Name
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetByListNoPagination(input *organizationModel.Field) (int, any) {
	output := &organizationModel.List{}
	quantity, organizationBase, err := m.OrganizationService.GetByListNoPagination(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	output.Total.Total = quantity
	organizationByte, err := json.Marshal(organizationBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = json.Unmarshal(organizationByte, &output.Organizations)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	for i, organization := range output.Organizations {
		organization.CreatedBy = *organizationBase[i].CreatedByUsers.Name
		organization.UpdatedBy = *organizationBase[i].UpdatedByUsers.Name
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetBySingle(input *organizationModel.Field) (int, any) {
	organizationBase, err := m.OrganizationService.GetBySingle(input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output := &organizationModel.Single{}
	organizationByte, _ := json.Marshal(organizationBase)
	err = json.Unmarshal(organizationByte, &output)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output.CreatedBy = *organizationBase.CreatedByUsers.Name
	output.UpdatedBy = *organizationBase.UpdatedByUsers.Name

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) Delete(input *organizationModel.Field) (int, any) {
	_, err := m.OrganizationService.GetBySingle(&organizationModel.Field{
		ID: input.ID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.OrganizationService.Delete(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, "Delete ok!")
}

func (m *manager) Update(input *organizationModel.Update) (int, any) {
	organizationBase, err := m.OrganizationService.GetBySingle(&organizationModel.Field{
		ID: input.ID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.OrganizationService.Update(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, organizationBase.ID)
}
