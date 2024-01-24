package department

import (
	"errors"
	"github.com/bytedance/sonic"
	affiliationModel "hta/internal/interactor/models/affiliations"
	resourceModel "hta/internal/interactor/models/resources"
	roleModel "hta/internal/interactor/models/roles"
	userModel "hta/internal/interactor/models/users"
	"hta/internal/interactor/pkg/util"
	affiliationService "hta/internal/interactor/service/affiliation"
	resourceService "hta/internal/interactor/service/resource"
	roleService "hta/internal/interactor/service/role"
	userService "hta/internal/interactor/service/user"
	"strings"

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
	Delete(trx *gorm.DB, input *departmentModel.Update) (int, any)
	Update(trx *gorm.DB, input *departmentModel.Update) (int, any)
}

type manager struct {
	DepartmentService  departmentService.Service
	AffiliationService affiliationService.Service
	RoleService        roleService.Service
	UserService        userService.Service
	ResourceService    resourceService.Service
}

func Init(db *gorm.DB) Manager {
	return &manager{
		DepartmentService:  departmentService.Init(db),
		AffiliationService: affiliationService.Init(db),
		RoleService:        roleService.Init(db),
		UserService:        userService.Init(db),
		ResourceService:    resourceService.Init(db),
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

	// get all users' info
	userBase, err := m.UserService.GetByListNoPagination(&userModel.Field{})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	var users []*userModel.Single
	userByte, err := sonic.Marshal(userBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(userByte, &users)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// create user map
	userMap := make(map[string]*userModel.Single)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// sync create supervisor
	if input.SupervisorID != nil {
		_, err = m.AffiliationService.WithTrx(trx).Create(&affiliationModel.Create{
			UserID:       *input.SupervisorID,
			DeptID:       *departmentBase.ID,
			IsSupervisor: true,
			JobTitle:     "部門主管",
			CreatedBy:    input.CreatedBy,
		})
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}

		// sync update the new supervisor's role
		if userMap[*input.SupervisorID].RoleID != roleMap["admin"] {
			err = m.UserService.WithTrx(trx).Update(&userModel.Update{
				ID:        *input.SupervisorID,
				RoleID:    util.PointerString(roleMap["supervisor"]),
				UpdatedBy: util.PointerString(input.CreatedBy),
			})
			if err != nil {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
		}

		// sync update the new supervisor's resource_group
		resourceBase, err := m.ResourceService.GetBySingle(&resourceModel.Field{
			ResourceUUID: userMap[*input.SupervisorID].ResourceUUID,
		})
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
		}

		if resourceBase != nil {
			if resourceBase.ResourceGroup != nil {
				// sync update resource_group
				newResGroup := strings.TrimSuffix(*resourceBase.ResourceGroup, "]") + `,"` + input.Name + `"]`
				err = m.ResourceService.WithTrx(trx).Update(&resourceModel.Update{
					ResourceUUID:  *resourceBase.ResourceUUID,
					ResourceGroup: util.PointerString(newResGroup),
					UpdatedBy:     util.PointerString(input.CreatedBy),
				})
				if err != nil {
					log.Error(err)
					return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
				}
			}
		}
	}

	// sync create affiliations
	if len(input.Affiliations) > 0 {
		var (
			affiliations []*affiliationModel.Create
			resUUIDs     []*string
		)
		for _, affiliation := range input.Affiliations {
			affiliations = append(affiliations, &affiliationModel.Create{
				UserID:    affiliation.UserID,
				DeptID:    *departmentBase.ID,
				JobTitle:  affiliation.JobTitle,
				CreatedBy: input.CreatedBy,
			})

			// collect resource uuid
			resUUIDs = append(resUUIDs, util.PointerString(userMap[affiliation.UserID].ResourceUUID))
		}

		_, err = m.AffiliationService.WithTrx(trx).CreateAll(affiliations)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}

		// get all affiliations' resource
		resourceBase, err := m.ResourceService.GetByListNoPagination(&resourceModel.Field{
			ResourceUUIDs: resUUIDs,
		})
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
		}

		// add new resource group
		if resourceBase != nil {
			for _, resource := range resourceBase {
				if resource.ResourceGroup != nil {
					// sync update resource_group
					newResGroup := strings.TrimSuffix(*resource.ResourceGroup, "]") + `,"` + input.Name + `"]`
					err = m.ResourceService.WithTrx(trx).Update(&resourceModel.Update{
						ResourceUUID:  *resource.ResourceUUID,
						ResourceGroup: util.PointerString(newResGroup),
						UpdatedBy:     util.PointerString(input.CreatedBy),
					})
					if err != nil {
						log.Error(err)
						return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
					}
				}
			}
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

func (m *manager) Delete(trx *gorm.DB, input *departmentModel.Update) (int, any) {
	defer trx.Rollback()

	deptBase, err := m.DepartmentService.GetBySingle(&departmentModel.Field{
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

	// get all affiliations of the department
	affiliationBase, err := m.AffiliationService.GetByListNoPagination(&affiliationModel.Field{
		DeptID: util.PointerString(input.ID),
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// collect resource uuid
	var resUUIDs []*string
	for _, affiliation := range affiliationBase {
		resUUIDs = append(resUUIDs, util.PointerString(*affiliation.Users.ResourceUUID))
	}

	// get all affiliations' resource
	resourceBase, err := m.ResourceService.GetByListNoPagination(&resourceModel.Field{
		ResourceUUIDs: resUUIDs,
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// delete resource_group
	if resourceBase != nil {
		for _, resource := range resourceBase {
			if resource.ResourceGroup != nil {
				// sync update resource_group
				newResGroup := strings.Replace(*resource.ResourceGroup, `,"`+*deptBase.Name+`"`, "", -1)
				err = m.ResourceService.WithTrx(trx).Update(&resourceModel.Update{
					ResourceUUID:  *resource.ResourceUUID,
					ResourceGroup: util.PointerString(newResGroup),
					UpdatedBy:     input.UpdatedBy,
				})
				if err != nil {
					log.Error(err)
					return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
				}
			}
		}
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

	if input.Name != nil {
		if *input.Name != *departmentBase.Name {
			// get all affiliations of the department
			affiliationBase, err := m.AffiliationService.GetByListNoPagination(&affiliationModel.Field{
				DeptID: util.PointerString(input.ID),
			})
			if err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					log.Error(err)
					return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
				}
			}

			// collect resource uuid
			var resUUIDs []*string
			for _, affiliation := range affiliationBase {
				resUUIDs = append(resUUIDs, util.PointerString(*affiliation.Users.ResourceUUID))
			}

			// get all affiliations' resource
			resourceBase, err := m.ResourceService.GetByListNoPagination(&resourceModel.Field{
				ResourceUUIDs: resUUIDs,
			})
			if err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					log.Error(err)
					return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
				}
			}

			// update resource_group
			if resourceBase != nil {
				for _, resource := range resourceBase {
					if resource.ResourceGroup != nil {
						// sync update resource_group
						newResGroup := strings.Replace(*resource.ResourceGroup, `,"`+*departmentBase.Name+`"`, `,"`+*input.Name+`"`, -1)
						err = m.ResourceService.WithTrx(trx).Update(&resourceModel.Update{
							ResourceUUID:  *resource.ResourceUUID,
							ResourceGroup: util.PointerString(newResGroup),
							UpdatedBy:     input.UpdatedBy,
						})
						if err != nil {
							log.Error(err)
							return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
						}
					}
				}
			}
		}
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

			// sync update the original supervisor's affiliation
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

	if input.Affiliations != nil {
		// get original affiliations of the department
		affiliationBase, err := m.AffiliationService.GetByListNoPagination(&affiliationModel.Field{
			DeptID: util.PointerString(input.ID),
		})
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
		}

		// create original affiliation map
		originalMap := make(map[string]bool)
		for _, affiliation := range affiliationBase {
			originalMap[*affiliation.UserID] = true
		}

		// create new affiliation map
		newMap := make(map[string]bool)
		for _, affiliation := range input.Affiliations {
			newMap[*affiliation.UserID] = true
		}

		// get all users' info
		userBase, err := m.UserService.GetByListNoPagination(&userModel.Field{})
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
		}

		// create user's resource_uuid map
		userMap := make(map[string]string)
		for _, user := range userBase {
			userMap[*user.ID] = *user.ResourceUUID
		}

		// determine if the affiliation in the new map but not in the original map, then create
		var newResUUID []*string
		for _, affiliation := range input.Affiliations {
			if _, ok := originalMap[*affiliation.UserID]; !ok {
				var jobTitle string
				if affiliation.JobTitle != nil {
					jobTitle = *affiliation.JobTitle
				}
				_, err = m.AffiliationService.WithTrx(trx).Create(&affiliationModel.Create{
					UserID:    *affiliation.UserID,
					DeptID:    input.ID,
					JobTitle:  jobTitle,
					CreatedBy: *input.UpdatedBy,
				})
				if err != nil {
					log.Error(err)
					return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
				}

				// collect resource uuid
				newResUUID = append(newResUUID, util.PointerString(userMap[*affiliation.UserID]))
			}
		}

		// determine if the affiliation in the original map but not in the new map, then delete
		var originalResUUID []*string
		for _, affiliation := range affiliationBase {
			if _, ok := newMap[*affiliation.UserID]; !ok {
				err = m.AffiliationService.WithTrx(trx).Delete(&affiliationModel.Field{
					ID: *affiliation.ID,
				})
				if err != nil {
					log.Error(err)
					return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
				}

				// collect resource uuid
				originalResUUID = append(originalResUUID, util.PointerString(userMap[*affiliation.UserID]))
			}
		}

		// sync update new resource_group
		if len(newResUUID) > 0 {
			// get all affiliations' resource
			resourceBase, err := m.ResourceService.GetByListNoPagination(&resourceModel.Field{
				ResourceUUIDs: newResUUID,
			})
			if err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					log.Error(err)
					return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
				}
			}

			if resourceBase != nil {
				for _, resource := range resourceBase {
					if resource.ResourceGroup != nil {
						// sync update resource_group
						newResGroup := strings.TrimSuffix(*resource.ResourceGroup, "]") + `,"` + *departmentBase.Name + `"]`
						err = m.ResourceService.WithTrx(trx).Update(&resourceModel.Update{
							ResourceUUID:  *resource.ResourceUUID,
							ResourceGroup: util.PointerString(newResGroup),
							UpdatedBy:     input.UpdatedBy,
						})
						if err != nil {
							log.Error(err)
							return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
						}
					}
				}
			}
		}

		// sync update original resource_group
		if len(originalResUUID) > 0 {
			// get all affiliations' resource
			resourceBase, err := m.ResourceService.GetByListNoPagination(&resourceModel.Field{
				ResourceUUIDs: originalResUUID,
			})
			if err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					log.Error(err)
					return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
				}
			}

			if resourceBase != nil {
				for _, resource := range resourceBase {
					if resource.ResourceGroup != nil {
						// sync update resource_group
						newResGroup := strings.Replace(*resource.ResourceGroup, `,"`+*departmentBase.Name+`"`, "", -1)
						err = m.ResourceService.WithTrx(trx).Update(&resourceModel.Update{
							ResourceUUID:  *resource.ResourceUUID,
							ResourceGroup: util.PointerString(newResGroup),
							UpdatedBy:     input.UpdatedBy,
						})
						if err != nil {
							log.Error(err)
							return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
						}
					}
				}
			}
		}
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, departmentBase.ID)
}
