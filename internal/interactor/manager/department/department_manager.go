package department

import (
	"errors"
	"github.com/bytedance/sonic"
	affiliationModel "hta/internal/interactor/models/affiliations"
	roleModel "hta/internal/interactor/models/roles"
	userModel "hta/internal/interactor/models/users"
	"hta/internal/interactor/pkg/util"
	affiliationService "hta/internal/interactor/service/affiliation"
	roleService "hta/internal/interactor/service/role"
	userService "hta/internal/interactor/service/user"

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
	RoleService        roleService.Service
	UserService        userService.Service
}

func Init(db *gorm.DB) Manager {
	return &manager{
		DepartmentService:  departmentService.Init(db),
		AffiliationService: affiliationService.Init(db),
		RoleService:        roleService.Init(db),
		UserService:        userService.Init(db),
	}
}

func (m *manager) Create(trx *gorm.DB, input *departmentModel.Create) (int, any) {
	defer trx.Rollback()

	// check the department exist
	quantity, err := m.DepartmentService.GetByQuantity(&departmentModel.Field{
		Name: util.PointerString(input.Name),
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	if quantity > 0 {
		return code.BadRequest, code.GetCodeMessage(code.BadRequest, "Department already exists!")
	}

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

	// get department's supervisor
	userBase, err := m.AffiliationService.GetByListNoPagination(&affiliationModel.Field{
		IsSupervisor: util.PointerBool(true),
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// create department's supervisor map
	userMap := make(map[string]string)
	for _, user := range userBase {
		userMap[*user.DeptID] = *user.Users.Name
	}

	for i, department := range output.Departments {
		department.CreatedBy = *departmentBase[i].CreatedByUsers.Name
		department.UpdatedBy = *departmentBase[i].UpdatedByUsers.Name
		department.Supervisor = userMap[department.ID]
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

	// get department's supervisor
	userBase, err := m.AffiliationService.GetByListNoPagination(&affiliationModel.Field{
		IsSupervisor: util.PointerBool(true),
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// create department's supervisor map
	userMap := make(map[string]string)
	for _, user := range userBase {
		userMap[*user.DeptID] = *user.Users.Name
	}

	for i, department := range output.Departments {
		department.CreatedBy = *departmentBase[i].CreatedByUsers.Name
		department.UpdatedBy = *departmentBase[i].UpdatedByUsers.Name
		department.Supervisor = userMap[department.ID]
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

	// get department's supervisor
	userBase, err := m.AffiliationService.GetBySingle(&affiliationModel.Field{
		DeptID:       util.PointerString(input.ID),
		IsSupervisor: util.PointerBool(true),
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	if userBase != nil {
		output.Supervisor = *userBase.Users.Name
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
		// get all roles
		roleBase, err := m.RoleService.GetByListNoPagination(&roleModel.Field{})
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
		}

		// create role map
		roleMap := make(map[string]string)
		for _, role := range roleBase {
			roleMap[*role.Name] = *role.ID
		}

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
			// check if the original supervisor is other department's supervisor
			quantity, err := m.AffiliationService.GetByQuantity(&affiliationModel.Field{
				UserID:       originalAffiliationBase.UserID,
				IsSupervisor: util.PointerBool(true),
			})
			if err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					log.Error(err)
					return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
				}
			}

			// get the original supervisor's user info
			userBase, err := m.UserService.GetBySingle(&userModel.Field{
				ID: *originalAffiliationBase.UserID,
			})
			if err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					log.Error(err)
					return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
				}
			}

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

			// sync update the original supervisor's role
			if quantity <= 1 && *userBase.RoleID != roleMap["admin"] {
				err = m.UserService.WithTrx(trx).Update(&userModel.Update{
					ID:        *originalAffiliationBase.UserID,
					RoleID:    util.PointerString(roleMap["user"]),
					UpdatedBy: input.UpdatedBy,
				})
				if err != nil {
					log.Error(err)
					return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
				}
			}
		}

		// get the new supervisor's user info
		userBase, err := m.UserService.GetBySingle(&userModel.Field{
			ID: *input.SupervisorID,
		})
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
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

		// sync update the new supervisor's role
		if *userBase.RoleID != roleMap["admin"] {
			err = m.UserService.WithTrx(trx).Update(&userModel.Update{
				ID:        *input.SupervisorID,
				RoleID:    util.PointerString(roleMap["supervisor"]),
				UpdatedBy: input.UpdatedBy,
			})
			if err != nil {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
		}
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, departmentBase.ID)
}
