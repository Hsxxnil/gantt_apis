package user

import (
	"errors"
	"github.com/bytedance/sonic"
	affiliationModel "hta/internal/interactor/models/affiliations"
	departmentModel "hta/internal/interactor/models/departments"
	resourceModel "hta/internal/interactor/models/resources"
	"hta/internal/interactor/pkg/util"
	affiliationService "hta/internal/interactor/service/affiliation"
	departmentService "hta/internal/interactor/service/department"
	resourceService "hta/internal/interactor/service/resource"

	userModel "hta/internal/interactor/models/users"
	userService "hta/internal/interactor/service/user"

	"gorm.io/gorm"

	"hta/internal/interactor/pkg/util/code"
	"hta/internal/interactor/pkg/util/log"
)

type Manager interface {
	GetByList(input *userModel.Fields) (int, any)
	GetByListNoPagination(input *userModel.Field) (int, any)
	GetBySingle(input *userModel.Field) (int, any)
	Delete(trx *gorm.DB, input *userModel.Update) (int, any)
	Update(trx *gorm.DB, input *userModel.Update) (int, any)
	Enable(input *userModel.Enable) (int, any)
	ResetPassword(input *userModel.ResetPassword) (int, any)
	Duplicate(input *userModel.Field) (int, any)
}

type manager struct {
	UserService        userService.Service
	AffiliationService affiliationService.Service
	DepartmentService  departmentService.Service
	ResourceService    resourceService.Service
}

func Init(db *gorm.DB) Manager {
	return &manager{
		UserService:        userService.Init(db),
		AffiliationService: affiliationService.Init(db),
		DepartmentService:  departmentService.Init(db),
		ResourceService:    resourceService.Init(db),
	}
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
	userByte, err := sonic.Marshal(userBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output.Pages = util.Pagination(quantity, output.Limit)
	err = sonic.Unmarshal(userByte, &output.Users)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// collect user IDs for efficient batch processing
	var userIds []*string
	for _, user := range userBase {
		userIds = append(userIds, user.ID)
	}

	// get all job titles
	_, affiliationBase, err := m.AffiliationService.GetByListNoPagination(&affiliationModel.Field{
		UserIDs: userIds,
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// transform affiliationBase to affiliations.Single
	affiliations := make([]*affiliationModel.Single, len(affiliationBase))
	affiliationByte, _ := sonic.Marshal(affiliationBase)
	err = sonic.Unmarshal(affiliationByte, &affiliations)

	// build maps for efficient lookups and collect department IDs
	affiliationMap := make(map[string]*affiliationModel.Single)
	var deptIds []*string
	for _, affiliation := range affiliations {
		affiliationMap[affiliation.UserID] = affiliation
		deptIds = append(deptIds, util.PointerString(affiliation.DeptID))
	}

	// get all departments
	_, departmentBase, err := m.DepartmentService.GetByListNoPagination(&departmentModel.Field{
		DeptIDs: deptIds,
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// transform departmentBase to departments.Single
	departments := make([]*departmentModel.Single, len(departmentBase))
	departmentByte, _ := sonic.Marshal(departmentBase)
	err = sonic.Unmarshal(departmentByte, &departments)

	// build maps for efficient lookups
	deptMap := make(map[string]*departmentModel.Single)
	for _, dept := range departments {
		deptMap[dept.ID] = dept
	}

	// assign job title and department to each user
	for i, user := range output.Users {
		user.Role = *userBase[i].Roles.DisplayName

		// get the user's job title and department
		if affiliation, ok := affiliationMap[*userBase[i].ID]; ok {
			user.JobTitle = affiliation.JobTitle
			if dept, ok := deptMap[affiliationMap[*userBase[i].ID].DeptID]; ok {
				user.Dept = dept.Name
			}
		}
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
	userByte, err := sonic.Marshal(userBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	err = sonic.Unmarshal(userByte, &output.Users)
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
	userByte, _ := sonic.Marshal(userBase)
	err = sonic.Unmarshal(userByte, &output)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output.Role = *userBase.Roles.DisplayName

	// get the user's job title
	affiliationBase, err := m.AffiliationService.GetBySingle(&affiliationModel.Field{
		UserID: util.PointerString(input.ID),
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	if affiliationBase != nil {
		if affiliationBase.JobTitle != nil {
			output.JobTitle = *affiliationBase.JobTitle
		}

		// get the user's department
		departmentBase, err := m.DepartmentService.GetBySingle(&departmentModel.Field{
			ID: *affiliationBase.DeptID,
		})
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
		}
		output.Dept = *departmentBase.Name
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) Delete(trx *gorm.DB, input *userModel.Update) (int, any) {
	defer trx.Rollback()

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

	err = m.UserService.WithTrx(trx).Delete(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// sync delete affiliation
	err = m.AffiliationService.WithTrx(trx).Delete(&affiliationModel.Field{
		UserID: util.PointerString(input.ID),
	})
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, "Delete ok!")
}

func (m *manager) Update(trx *gorm.DB, input *userModel.Update) (int, any) {
	defer trx.Rollback()

	// validate old password
	if input.Password != nil {
		acknowledge, _, err := m.UserService.AcknowledgeUser(&userModel.Field{
			ID:       input.ID,
			Password: input.OldPassword,
		})
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
		}

		if acknowledge == false {
			return code.PermissionDenied, code.GetCodeMessage(code.PermissionDenied, "Incorrect password.")
		}
	}

	// determine if the username is duplicate
	if input.UserName != nil {
		quantity, _ := m.UserService.GetByQuantity(&userModel.Field{
			UserName: input.UserName,
		})

		if quantity > 0 {
			log.Info("UserName already exists. UserName: ", input.UserName)
			return code.BadRequest, code.GetCodeMessage(code.BadRequest, "User already exists.")
		}
	}

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

	// if the user has no resource, bind the resource to the user
	var resourceID *string
	if userBase.ResourceUUID == nil {
		// check if the resource with the same email exists
		resourceBase, err := m.ResourceService.GetBySingle(&resourceModel.Field{
			Email: userBase.Email,
		})
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
		}

		if resourceBase != nil {
			resourceID = resourceBase.ResourceUUID
		} else {
			// sync create resource
			newResourceBase, err := m.ResourceService.WithTrx(trx).Create(&resourceModel.Create{
				ResourceName: *userBase.Name,
				Email:        *userBase.Email,
				CreatedBy:    *input.UpdatedBy,
			})
			if err != nil {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
			resourceID = newResourceBase.ResourceUUID
		}
		input.ResourceUUID = resourceID
	}

	// sync delete affiliation
	err = m.AffiliationService.WithTrx(trx).Delete(&affiliationModel.Field{
		UserID: userBase.ID,
	})
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// sync create affiliation
	for _, affiliation := range input.Affiliations {
		_, err = m.AffiliationService.WithTrx(trx).Create(&affiliationModel.Create{
			UserID:    *userBase.ID,
			DeptID:    affiliation.DeptID,
			JobTitle:  affiliation.JobTitle,
			CreatedBy: *input.UpdatedBy,
		})
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	err = m.UserService.WithTrx(trx).Update(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, userBase.ID)
}

func (m *manager) Enable(input *userModel.Enable) (int, any) {
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

	// set default value
	if input.IsEnabled == nil {
		input.IsEnabled = util.PointerBool(true)
	}

	// transform input to Update struct
	update := &userModel.Update{}
	inputByte, err := sonic.Marshal(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(inputByte, &update)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.UserService.Update(update)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, "Enable ok!")
}

func (m *manager) ResetPassword(input *userModel.ResetPassword) (int, any) {
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

	// transform input to Update struct
	update := &userModel.Update{}
	inputByte, err := sonic.Marshal(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(inputByte, update)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.UserService.Update(update)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, userBase.ID)
}

func (m *manager) Duplicate(input *userModel.Field) (int, any) {
	output := &userModel.IsDuplicate{}
	quantity, err := m.UserService.GetByQuantity(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	if quantity > 0 {
		log.Info("User already exists. UserName: ", input.FilterUserName, "email: ", input.FilterEmail)
		output.IsDuplicate = true
	} else {
		output.IsDuplicate = false
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}
