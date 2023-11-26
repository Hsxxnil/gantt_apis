package work_day

import (
	"encoding/json"
	"errors"
	"hta/internal/interactor/pkg/util"

	"gorm.io/gorm"

	workDayModel "hta/internal/interactor/models/work_days"
	workDayService "hta/internal/interactor/service/work_day"

	"hta/internal/interactor/pkg/util/code"
	"hta/internal/interactor/pkg/util/log"
)

type Manager interface {
	Create(trx *gorm.DB, input *workDayModel.Create) (int, any)
	GetByList(input *workDayModel.Fields) (int, any)
	GetByListNoPagination(input *workDayModel.Field) (int, any)
	GetBySingle(input *workDayModel.Field) (int, any)
	Delete(input *workDayModel.Field) (int, any)
	Update(input *workDayModel.Update) (int, any)
}

type manager struct {
	WorkDayService workDayService.Service
}

func Init(db *gorm.DB) Manager {
	return &manager{
		WorkDayService: workDayService.Init(db),
	}
}

func (m *manager) Create(trx *gorm.DB, input *workDayModel.Create) (int, any) {
	defer trx.Rollback()

	// 將workWeek轉換為JSON
	if len(input.WorkWeeks) > 0 {
		weekJson, _ := json.Marshal(input.WorkWeeks)
		input.WorkWeek = string(weekJson)
	}

	// 將workingTime轉換為JSON
	if len(input.WorkingTimes) > 0 {
		timeJson, _ := json.Marshal(input.WorkingTimes)
		input.WorkingTime = string(timeJson)
	}

	workDayBase, err := m.WorkDayService.WithTrx(trx).Create(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, workDayBase.ID)
}

func (m *manager) GetByList(input *workDayModel.Fields) (int, any) {
	output := &workDayModel.List{}
	output.Limit = input.Limit
	output.Page = input.Page
	quantity, workDayBase, err := m.WorkDayService.GetByList(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	output.Total.Total = quantity
	output.Pages = util.Pagination(quantity, output.Limit)
	workDayByte, err := json.Marshal(workDayBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = json.Unmarshal(workDayByte, &output.WorkDays)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	for i, work := range output.WorkDays {
		work.CreatedBy = *workDayBase[i].CreatedByUsers.Name
		work.UpdatedBy = *workDayBase[i].UpdatedByUsers.Name

		// 將work_week轉為陣列
		var workWeeks []string
		if *workDayBase[i].WorkWeek != "" {
			err = json.Unmarshal([]byte(*workDayBase[i].WorkWeek), &workWeeks)
			if err != nil {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
		}
		work.WorkWeeks = workWeeks

		// 將working_time轉為陣列
		var workingTimes []workDayModel.WorkingTimes
		if *workDayBase[i].WorkingTime != "" {
			err = json.Unmarshal([]byte(*workDayBase[i].WorkingTime), &workingTimes)
			if err != nil {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
		}
		work.WorkingTimes = workingTimes
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetByListNoPagination(input *workDayModel.Field) (int, any) {
	output := &workDayModel.List{}
	quantity, workDayBase, err := m.WorkDayService.GetByListNoPagination(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	output.Total.Total = quantity
	workDayByte, err := json.Marshal(workDayBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = json.Unmarshal(workDayByte, &output.WorkDays)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	for i, work := range output.WorkDays {
		work.CreatedBy = *workDayBase[i].CreatedByUsers.Name
		work.UpdatedBy = *workDayBase[i].UpdatedByUsers.Name

		// 將work_week轉為陣列
		var workWeeks []string
		if *workDayBase[i].WorkWeek != "" {
			err = json.Unmarshal([]byte(*workDayBase[i].WorkWeek), &workWeeks)
			if err != nil {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
		}
		work.WorkWeeks = workWeeks

		// 將working_time轉為陣列
		var workingTimes []workDayModel.WorkingTimes
		if *workDayBase[i].WorkingTime != "" {
			err = json.Unmarshal([]byte(*workDayBase[i].WorkingTime), &workingTimes)
			if err != nil {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
		}
		work.WorkingTimes = workingTimes
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetBySingle(input *workDayModel.Field) (int, any) {
	workDayBase, err := m.WorkDayService.GetBySingle(input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output := &workDayModel.Single{}
	workDayByte, _ := json.Marshal(workDayBase)
	err = json.Unmarshal(workDayByte, &output)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output.CreatedBy = *workDayBase.CreatedByUsers.Name
	output.UpdatedBy = *workDayBase.UpdatedByUsers.Name

	// 將work_week轉為陣列
	var workWeeks []string
	if *workDayBase.WorkWeek != "" {
		err = json.Unmarshal([]byte(*workDayBase.WorkWeek), &workWeeks)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}
	output.WorkWeeks = workWeeks

	// 將working_time轉為陣列
	var workingTimes []workDayModel.WorkingTimes
	if *workDayBase.WorkingTime != "" {
		err = json.Unmarshal([]byte(*workDayBase.WorkingTime), &workingTimes)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}
	output.WorkingTimes = workingTimes

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) Delete(input *workDayModel.Field) (int, any) {
	_, err := m.WorkDayService.GetBySingle(&workDayModel.Field{
		ID: input.ID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.WorkDayService.Delete(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, "Delete ok!")
}

func (m *manager) Update(input *workDayModel.Update) (int, any) {
	workDayBase, err := m.WorkDayService.GetBySingle(&workDayModel.Field{
		ID: input.ID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// 將workWeek轉換為JSON
	if len(input.WorkWeeks) > 0 {
		weekJson, _ := json.Marshal(input.WorkWeeks)
		input.WorkWeek = util.PointerString(string(weekJson))
	}

	// 將workingTime轉換為JSON
	if len(input.WorkingTimes) > 0 {
		timeJson, _ := json.Marshal(input.WorkingTimes)
		input.WorkingTime = util.PointerString(string(timeJson))
	}

	err = m.WorkDayService.Update(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, workDayBase.ID)
}
