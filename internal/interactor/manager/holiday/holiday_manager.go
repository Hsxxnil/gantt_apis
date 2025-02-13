package holiday

import (
	"errors"
	"gantt/internal/interactor/pkg/util"

	"github.com/bytedance/sonic"

	"gorm.io/gorm"

	holidayModel "gantt/internal/interactor/models/holidays"
	holidayService "gantt/internal/interactor/service/holiday"

	"gantt/internal/interactor/pkg/util/code"
	"gantt/internal/interactor/pkg/util/log"
)

type Manager interface {
	Create(trx *gorm.DB, input *holidayModel.Create) (int, any)
	GetByList(input *holidayModel.Fields) (int, any)
	GetByListNoPagination(input *holidayModel.Field) (int, any)
	GetBySingle(input *holidayModel.Field) (int, any)
	Delete(input *holidayModel.Field) (int, any)
	Update(input *holidayModel.Update) (int, any)
}

type manager struct {
	HolidayService holidayService.Service
}

func Init(db *gorm.DB) Manager {
	return &manager{
		HolidayService: holidayService.Init(db),
	}
}

func (m *manager) Create(trx *gorm.DB, input *holidayModel.Create) (int, any) {
	defer trx.Rollback()

	holidayBase, err := m.HolidayService.WithTrx(trx).Create(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, holidayBase.ID)
}

func (m *manager) GetByList(input *holidayModel.Fields) (int, any) {
	output := &holidayModel.List{}
	output.Limit = input.Limit
	output.Page = input.Page
	quantity, holidayBase, err := m.HolidayService.GetByList(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	output.Total.Total = quantity
	output.Pages = util.Pagination(quantity, output.Limit)
	holidayByte, err := sonic.Marshal(holidayBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(holidayByte, &output.Holidays)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	for i, holiday := range output.Holidays {
		holiday.CreatedBy = *holidayBase[i].CreatedByUsers.Name
		holiday.UpdatedBy = *holidayBase[i].UpdatedByUsers.Name
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetByListNoPagination(input *holidayModel.Field) (int, any) {
	output := &holidayModel.List{}
	holidayBase, err := m.HolidayService.GetByListNoPagination(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	holidayByte, err := sonic.Marshal(holidayBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(holidayByte, &output.Holidays)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	for i, holiday := range output.Holidays {
		holiday.CreatedBy = *holidayBase[i].CreatedByUsers.Name
		holiday.UpdatedBy = *holidayBase[i].UpdatedByUsers.Name
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetBySingle(input *holidayModel.Field) (int, any) {
	holidayBase, err := m.HolidayService.GetBySingle(input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output := &holidayModel.Single{}
	holidayByte, _ := sonic.Marshal(holidayBase)
	err = sonic.Unmarshal(holidayByte, &output)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output.CreatedBy = *holidayBase.CreatedByUsers.Name
	output.UpdatedBy = *holidayBase.UpdatedByUsers.Name

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) Delete(input *holidayModel.Field) (int, any) {
	_, err := m.HolidayService.GetBySingle(&holidayModel.Field{
		ID: input.ID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.HolidayService.Delete(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, "Delete ok!")
}

func (m *manager) Update(input *holidayModel.Update) (int, any) {
	holidayBase, err := m.HolidayService.GetBySingle(&holidayModel.Field{
		ID: input.ID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.HolidayService.Update(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, holidayBase.ID)
}
