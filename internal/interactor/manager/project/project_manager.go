package project

import (
	"errors"
	eventMarkModel "gantt/internal/interactor/models/event_marks"
	projectResourceModel "gantt/internal/interactor/models/project_resources"
	projectTypeModel "gantt/internal/interactor/models/project_types"
	resourceModel "gantt/internal/interactor/models/resources"
	roleModel "gantt/internal/interactor/models/roles"
	taskResourceModel "gantt/internal/interactor/models/task_resources"
	taskModel "gantt/internal/interactor/models/tasks"
	userModel "gantt/internal/interactor/models/users"
	"gantt/internal/interactor/pkg/util"
	eventMarkService "gantt/internal/interactor/service/event_mark"
	projectResourceService "gantt/internal/interactor/service/project_resource"
	projectTypeService "gantt/internal/interactor/service/project_type"
	resourceService "gantt/internal/interactor/service/resource"
	roleService "gantt/internal/interactor/service/role"
	taskService "gantt/internal/interactor/service/task"
	taskResourceService "gantt/internal/interactor/service/task_resource"
	userService "gantt/internal/interactor/service/user"
	"time"

	"github.com/bytedance/sonic"

	"gorm.io/gorm"

	projectModel "gantt/internal/interactor/models/projects"
	projectService "gantt/internal/interactor/service/project"

	"gantt/internal/interactor/pkg/util/code"
	"gantt/internal/interactor/pkg/util/log"
)

type Manager interface {
	Create(trx *gorm.DB, input *projectModel.Create) (int, any)
	GetByList(input *projectModel.Fields) (int, any)
	GetByListNoPagination(input *projectModel.Field) (int, any)
	GetBySingle(input *projectModel.Field) (int, any)
	Delete(trx *gorm.DB, input *projectModel.Update) (int, any)
	Update(trx *gorm.DB, input *projectModel.Update) (int, any)
}

type manager struct {
	ProjectService         projectService.Service
	TaskService            taskService.Service
	ResourceService        resourceService.Service
	ProjectTypeService     projectTypeService.Service
	ProjectResourceService projectResourceService.Service
	EventMarkService       eventMarkService.Service
	RoleService            roleService.Service
	TaskResourceService    taskResourceService.Service
	UserService            userService.Service
}

func Init(db *gorm.DB) Manager {
	return &manager{
		ProjectService:         projectService.Init(db),
		TaskService:            taskService.Init(db),
		ResourceService:        resourceService.Init(db),
		ProjectTypeService:     projectTypeService.Init(db),
		ProjectResourceService: projectResourceService.Init(db),
		EventMarkService:       eventMarkService.Init(db),
		RoleService:            roleService.Init(db),
		TaskResourceService:    taskResourceService.Init(db),
		UserService:            userService.Init(db),
	}
}

func (m *manager) Create(trx *gorm.DB, input *projectModel.Create) (int, any) {
	defer trx.Rollback()

	projectBase, err := m.ProjectService.WithTrx(trx).Create(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// sync create project_resource
	var resourceList []*projectResourceModel.Create
	if len(input.Resource) > 0 {
		for _, resource := range input.Resource {
			proRes := &projectResourceModel.Create{
				ProjectUUID:  *projectBase.ProjectUUID,
				ResourceUUID: resource.ResourceUUID,
				Role:         resource.Role,
				IsEditable:   true,
				CreatedBy:    input.CreatedBy,
			}
			resourceList = append(resourceList, proRes)
		}

		_, err := m.ProjectResourceService.WithTrx(trx).CreateAll(resourceList)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, projectBase.ProjectUUID)
}

func (m *manager) GetByList(input *projectModel.Fields) (int, any) {
	output := &projectModel.List{}
	output.Limit = input.Limit
	output.Page = input.Page
	// search project type
	if input.FilterType != nil {
		projectTypeBase, _ := m.ProjectTypeService.GetByListNoPagination(&projectTypeModel.Field{
			Names: input.FilterType,
		})
		if len(projectTypeBase) > 0 {
			for _, projectType := range projectTypeBase {
				input.FilterTypes = append(input.FilterTypes, projectType.ID)
			}
		} else {
			input.FilterTypes = nil
		}
	}

	// search manager
	if input.FilterManager != "" {
		resourceBase, _ := m.ResourceService.GetByListNoPagination(&resourceModel.Field{
			ResourceName: util.PointerString(input.FilterManager),
		})
		if len(resourceBase) > 0 {
			for _, resource := range resourceBase {
				input.FilterManagers = append(input.FilterManagers, resource.ResourceUUID)
			}
		}

		// search project_resource
		proResBase, _ := m.ProjectResourceService.GetByListNoPagination(&projectResourceModel.Field{
			ResourceUUIDs: input.FilterManagers,
			Role:          util.PointerString("PM"),
		})
		if len(proResBase) > 0 {
			for _, proRes := range proResBase {
				input.ProjectUUIDs = append(input.ProjectUUIDs, proRes.ProjectUUID)
			}
		}
	}

	// if the user is user, search the project which is created by the user or the user is the project's member
	if *input.Role == "user" {
		input.CreatedBy = input.UserID
		// search project_resource
		proResBase, _ := m.ProjectResourceService.GetByListNoPagination(&projectResourceModel.Field{
			ResourceUUID: input.ResUUID,
		})

		if len(proResBase) > 0 {
			for _, proRes := range proResBase {
				input.ProjectUUIDs = append(input.ProjectUUIDs, proRes.ProjectUUID)
			}
		}
	}

	quantity, projectBase, err := m.ProjectService.GetByList(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	output.Total.Total = quantity
	output.Pages = util.Pagination(quantity, output.Limit)
	projectByte, err := sonic.Marshal(projectBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(projectByte, &output.Projects)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// get projects' manager
	proResBase, err := m.ProjectResourceService.GetByListNoPagination(&projectResourceModel.Field{
		Role: util.PointerString("PM"),
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// create project's manager map
	proResNameMap := make(map[string]string)
	proResUUIDMap := make(map[string]string)
	for _, proRes := range proResBase {
		proResNameMap[*proRes.ProjectUUID] = *proRes.Resources.ResourceName
		proResUUIDMap[*proRes.ProjectUUID] = *proRes.ResourceUUID
	}

	today := time.Now().UTC()
	for i, project := range output.Projects {
		project.Type = *projectBase[i].ProjectTypes.Name
		project.Manager = proResNameMap[*projectBase[i].ProjectUUID]
		project.CreatedBy = *projectBase[i].CreatedByUsers.Name
		project.UpdatedBy = *projectBase[i].UpdatedByUsers.Name
		// calculate project progress
		if projectBase[i].StartDate != nil && projectBase[i].EndDate != nil {
			progress := int64((today.Sub(*projectBase[i].StartDate).Hours() / projectBase[i].EndDate.Sub(*projectBase[i].StartDate).Hours()) * 100)
			if progress <= 0 {
				project.Progress = 0
			} else if progress >= 100 {
				project.Progress = 100
			} else {
				project.Progress = progress
			}
		} else {
			project.Progress = 0
		}

		// check the user can edit or delete the project
		if *input.Role == "admin" {
			project.IsEditable = true
		} else {
			if *projectBase[i].CreatedBy == *input.UserID || proResUUIDMap[*projectBase[i].ProjectUUID] == *input.ResUUID {
				project.IsEditable = true
			}
		}
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetByListNoPagination(input *projectModel.Field) (int, any) {
	output := &projectModel.ListNoPagination{}

	// if the user is user, search the project which is created by the user or the user is the project's member
	if *input.Role == "user" {
		input.CreatedBy = input.UserID
		// search project_resource
		proResBase, _ := m.ProjectResourceService.GetByListNoPagination(&projectResourceModel.Field{
			ResourceUUID: input.ResUUID,
		})
		if len(proResBase) > 0 {
			for _, proRes := range proResBase {
				input.ProjectUUIDs = append(input.ProjectUUIDs, proRes.ProjectUUID)
			}
		}
	}

	projectBase, err := m.ProjectService.GetByListNoPagination(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	projectByte, err := sonic.Marshal(projectBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(projectByte, &output.Projects)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetBySingle(input *projectModel.Field) (int, any) {
	// if the user is user, search the project which is created by the user or the user is the project's member
	if *input.Role == "user" {
		input.CreatedBy = input.UserID
		// search project_resource
		_, err := m.ProjectResourceService.GetBySingle(&projectResourceModel.Field{
			ResourceUUID: input.ResUUID,
			ProjectUUID:  util.PointerString(input.ProjectUUID),
		})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
			}

			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	projectBase, err := m.ProjectService.GetBySingle(input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output := &projectModel.Single{}
	projectByte, _ := sonic.Marshal(projectBase)
	err = sonic.Unmarshal(projectByte, &output)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// get projects' manager
	proResBase, err := m.ProjectResourceService.GetBySingle(&projectResourceModel.Field{
		ProjectUUID: util.PointerString(input.ProjectUUID),
		Role:        util.PointerString("PM"),
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	if proResBase != nil {
		output.Manager = *proResBase.Resources.ResourceName
	}
	output.Type = *projectBase.ProjectTypes.Name
	output.CreatedBy = *projectBase.CreatedByUsers.Name
	output.UpdatedBy = *projectBase.UpdatedByUsers.Name
	today := time.Now().UTC()
	if projectBase.StartDate != nil && projectBase.EndDate != nil {
		progress := int64((today.Sub(*projectBase.StartDate).Hours() / projectBase.EndDate.Sub(*projectBase.StartDate).Hours()) * 100)
		if progress <= 0 {
			output.Progress = 0
		} else if progress >= 100 {
			output.Progress = 100
		} else {
			output.Progress = progress
		}
	} else {
		output.Progress = 0
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) Delete(trx *gorm.DB,
	input *projectModel.Update) (int, any) {
	defer trx.Rollback()

	// check the update_by is the project's manager
	pmBase, err := m.ProjectResourceService.GetBySingle(&projectResourceModel.Field{
		ProjectUUID: util.PointerString(input.ProjectUUID),
		Role:        util.PointerString("PM"),
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// check the update_by is the project's creator
	projectBase, err := m.ProjectService.GetBySingle(&projectModel.Field{
		ProjectUUID: input.ProjectUUID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	if *input.Role != "admin" {
		if *projectBase.CreatedBy != *input.UpdatedBy {
			if pmBase != nil {
				if *pmBase.ResourceUUID != *input.ResUUID {
					log.Info("The user don't have permission to update this project.")
					return code.BadRequest, code.GetCodeMessage(code.BadRequest, "The user don't have permission to update this project.")
				}
			} else {
				log.Info("The user don't have permission to update this project.")
				return code.BadRequest, code.GetCodeMessage(code.BadRequest, "The user don't have permission to update this project.")
			}
		}
	}

	err = m.ProjectService.WithTrx(trx).Delete(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// sync delete task
	err = m.TaskService.WithTrx(trx).Delete(&taskModel.Field{
		ProjectUUID: util.PointerString(input.ProjectUUID),
	})
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// sync delete project_resource
	err = m.ProjectResourceService.WithTrx(trx).Delete(&projectResourceModel.Field{
		ProjectUUID: util.PointerString(input.ProjectUUID),
	})
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// sync delete event_mark
	err = m.EventMarkService.Delete(&eventMarkModel.Field{
		ProjectUUID: util.PointerString(input.ProjectUUID),
	})
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, "Delete ok!")
}

func (m *manager) Update(trx *gorm.DB,
	input *projectModel.Update) (int, any) {
	defer trx.Rollback()

	// check the update_by is the project's manager
	pmBase, err := m.ProjectResourceService.GetBySingle(&projectResourceModel.Field{
		ProjectUUID: util.PointerString(input.ProjectUUID),
		Role:        util.PointerString("PM"),
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// check the update_by is the project's creator
	projectBase, err := m.ProjectService.GetBySingle(&projectModel.Field{
		ProjectUUID: input.ProjectUUID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	if *input.Role != "admin" {
		if *projectBase.CreatedBy != *input.UpdatedBy {
			if pmBase != nil {
				if *pmBase.ResourceUUID != *input.ResUUID {
					log.Info("The user don't have permission to update this project.")
					return code.BadRequest, code.GetCodeMessage(code.BadRequest, "The user don't have permission to update this project.")
				}
			} else {
				log.Info("The user don't have permission to update this project.")
				return code.BadRequest, code.GetCodeMessage(code.BadRequest, "The user don't have permission to update this project.")
			}
		}
	}

	err = m.ProjectService.Update(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// update resource
	var (
		resourceList  []*projectResourceModel.Create
		resourceUUIDs []*string
		taskUUIDs     []*string
	)

	// sync delete project_resource
	err = m.ProjectResourceService.WithTrx(trx).Delete(&projectResourceModel.Field{
		ProjectUUID: util.PointerString(input.ProjectUUID),
	})
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	if len(input.Resource) > 0 {
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
			roleMap[*role.ID] = *role.Name
		}

		// get all users info
		userBase, err := m.UserService.GetByListNoPagination(&userModel.Field{})
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
		}

		// create user map
		userRoleMap := make(map[string]string)
		for _, user := range userBase {
			userRoleMap[*user.ResourceUUID] = *user.RoleID
		}

		// sync create project_resource
		for _, resource := range input.Resource {
			// determine the user's is_editable
			isEditable := false
			if roleMap[userRoleMap[resource.ResourceUUID]] == "admin" {
				isEditable = true
			} else {
				if resource.IsEditable {
					isEditable = true
				}
			}

			proRes := &projectResourceModel.Create{
				ProjectUUID:  *projectBase.ProjectUUID,
				ResourceUUID: resource.ResourceUUID,
				Role:         resource.Role,
				IsEditable:   isEditable,
				CreatedBy:    *input.UpdatedBy,
			}
			resourceList = append(resourceList, proRes)
			resourceUUIDs = append(resourceUUIDs, util.PointerString(resource.ResourceUUID))
		}

		_, err = m.ProjectResourceService.WithTrx(trx).CreateAll(resourceList)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// get task_uuids
	taskBase, err := m.TaskService.GetByListNoPagination(&taskModel.Field{
		ProjectUUID: util.PointerString(input.ProjectUUID),
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}
	if len(taskBase) > 0 {
		for _, task := range taskBase {
			taskUUIDs = append(taskUUIDs, task.TaskUUID)
		}

		// sync delete task_resource
		err = m.TaskResourceService.WithTrx(trx).Delete(&taskResourceModel.Field{
			TaskUUIDs:     taskUUIDs,
			ResourceUUIDs: resourceUUIDs,
		})
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, projectBase.ProjectUUID)
}
