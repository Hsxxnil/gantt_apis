package user

import (
	"encoding/json"
	"errors"

	"hta/internal/interactor/pkg/util"

	userModel "hta/internal/interactor/models/users"
	userService "hta/internal/interactor/service/user"

	"gorm.io/gorm"

	"hta/internal/interactor/pkg/util/code"
	"hta/internal/interactor/pkg/util/log"
)

type Manager interface {
	Create(trx *gorm.DB, input *userModel.Create) (int, any)
	GetByList(input *userModel.Fields) (int, any)
	GetByListNoPagination(input *userModel.Field) (int, any)
	GetBySingle(input *userModel.Field) (int, any)
	Delete(input *userModel.Update) (int, any)
	Update(input *userModel.Update) (int, any)
}

type manager struct {
	UserService userService.Service
}

func Init(db *gorm.DB) Manager {
	return &manager{
		UserService: userService.Init(db),
	}
}

func (m *manager) Create(trx *gorm.DB, input *userModel.Create) (int, any) {
	defer trx.Rollback()

	// determine if the username is duplicate
	quantity, _ := m.UserService.GetByQuantity(&userModel.Field{
		UserName: util.PointerString(input.UserName),
	})

	if quantity > 0 {
		log.Info("UserName already exists. UserName: ", input.UserName)
		return code.BadRequest, code.GetCodeMessage(code.BadRequest, "User already exists.")
	}

	userBase, err := m.UserService.WithTrx(trx).Create(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, userBase.ID)
}

func (m *manager) GetByList(input *userModel.Fields) (int, any) {
	output := &userModel.List{}
	output.Limit = input.Limit
	output.Page = input.Page
	quantity, userBase, err := m.UserService.GetByList(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	output.Total.Total = quantity
	userByte, err := json.Marshal(userBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	output.Pages = util.Pagination(quantity, output.Limit)
	err = json.Unmarshal(userByte, &output.Users)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	for i, user := range output.Users {
		user.Role = *userBase[i].Roles.DisplayName
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetByListNoPagination(input *userModel.Field) (int, any) {
	output := &userModel.ListNoPagination{}
	userBase, err := m.UserService.GetByListNoPagination(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	userByte, err := json.Marshal(userBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	err = json.Unmarshal(userByte, &output.Users)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetBySingle(input *userModel.Field) (int, any) {
	userBase, err := m.UserService.GetBySingle(input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output := &userModel.Single{}
	userByte, _ := json.Marshal(userBase)
	err = json.Unmarshal(userByte, &output)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output.Role = *userBase.Roles.DisplayName

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) Delete(input *userModel.Update) (int, any) {
	_, err := m.UserService.GetBySingle(&userModel.Field{
		ID: input.ID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.UserService.Delete(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, "Delete ok!")
}

func (m *manager) Update(input *userModel.Update) (int, any) {
	userBase, err := m.UserService.GetBySingle(&userModel.Field{
		ID: input.ID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.UserService.Update(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, userBase.ID)
}
