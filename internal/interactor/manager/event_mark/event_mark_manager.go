package event_mark

import (
	"errors"

	"gorm.io/gorm"

	eventMarkModel "hta/internal/interactor/models/event_marks"
	eventMarkService "hta/internal/interactor/service/event_mark"

	"hta/internal/interactor/pkg/util/code"
	"hta/internal/interactor/pkg/util/log"
)

type Manager interface {
	Create(trx *gorm.DB, input *eventMarkModel.Create) (int, any)
	Delete(input *eventMarkModel.Field) (int, any)
	Update(input *eventMarkModel.Update) (int, any)
}

type manager struct {
	EventMarkService eventMarkService.Service
}

func Init(db *gorm.DB) Manager {
	return &manager{
		EventMarkService: eventMarkService.Init(db),
	}
}

func (m *manager) Create(trx *gorm.DB, input *eventMarkModel.Create) (int, any) {
	defer trx.Rollback()

	eventMarkBase, err := m.EventMarkService.WithTrx(trx).Create(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, eventMarkBase.ID)
}

func (m *manager) Delete(input *eventMarkModel.Field) (int, any) {
	_, err := m.EventMarkService.GetBySingle(&eventMarkModel.Field{
		ID: input.ID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.EventMarkService.Delete(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, "Delete ok!")
}

func (m *manager) Update(input *eventMarkModel.Update) (int, any) {
	eventMarkBase, err := m.EventMarkService.GetBySingle(&eventMarkModel.Field{
		ID: input.ID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.EventMarkService.Update(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, eventMarkBase.ID)
}
