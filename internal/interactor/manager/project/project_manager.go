package project

import (
	"errors"
	"github.com/bytedance/sonic"
	eventMarkModel "hta/internal/interactor/models/event_marks"
	projectResourceModel "hta/internal/interactor/models/project_resources"
	projectTypeModel "hta/internal/interactor/models/project_types"
	resourceModel "hta/internal/interactor/models/resources"
	taskResourceModel "hta/internal/interactor/models/task_resources"
	taskModel "hta/internal/interactor/models/tasks"
	"hta/internal/interactor/pkg/util"
	eventMarkService "hta/internal/interactor/service/event_mark"
	projectResourceService "hta/internal/interactor/service/project_resource"
	projectTypeService "hta/internal/interactor/service/project_type"
	resourceService "hta/internal/interactor/service/resource"
	roleService "hta/internal/interactor/service/role"
	taskService "hta/internal/interactor/service/task"
	taskResourceService "hta/internal/interactor/service/task_resource"
	"time"

	"gorm.io/gorm"

	projectModel "hta/internal/interactor/models/projects"
	projectService "hta/internal/interactor/service/project"

	"hta/internal/interactor/pkg/util/code"
	"hta/internal/interactor/pkg/util/log"
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

	// if the user is not admin, search the project which is created by the user or the user is the project's member
	if input.CreatedBy != nil && input.ResourceUUID != nil {
		// search project_resource
		proResBase, _ := m.ProjectResourceService.GetByListNoPagination(&projectResourceModel.Field{
			ResourceUUID: input.ResourceUUID,
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
	projectMap := make(map[string]string)
	for _, proRes := range proResBase {
		projectMap[*proRes.ProjectUUID] = *proRes.Resources.ResourceName
	}

	today := time.Now().UTC()
	for i, project := range output.Projects {
		project.Type = *projectBase[i].ProjectTypes.Name
		project.Manager = projectMap[*projectBase[i].ProjectUUID]
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
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetByListNoPagination(input *projectModel.Field) (int, any) {
	output := &projectModel.List{}

	// if the user is not admin, search the project which is created by the user or the user is the project's member
	if input.CreatedBy != nil && input.ResourceUUID != nil {
		// search project_resource
		proResBase, _ := m.ProjectResourceService.GetByListNoPagination(&projectResourceModel.Field{
			ResourceUUID: input.ResourceUUID,
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
	projectMap := make(map[string]string)
	for _, proRes := range proResBase {
		projectMap[*proRes.ProjectUUID] = *proRes.Resources.ResourceName
	}

	today := time.Now().UTC()
	for i, project := range output.Projects {
		project.Type = *projectBase[i].ProjectTypes.Name
		project.Manager = projectMap[*projectBase[i].ProjectUUID]
		project.CreatedBy = *projectBase[i].CreatedByUsers.Name
		project.UpdatedBy = *projectBase[i].UpdatedByUsers.Name
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
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetBySingle(input *projectModel.Field) (int, any) {
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

func (m *manager) Delete(trx *gorm.DB, input *projectModel.Update) (int, any) {
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

	if *input.UpdateRole != "admin" {
		if *projectBase.CreatedBy != *input.UpdatedBy {
			if pmBase != nil {
				if *pmBase.ResourceUUID != *input.UpdateResUUID {
					return code.BadRequest, code.GetCodeMessage(code.BadRequest, "You don't have permission to delete this project!")
				}
			} else {
				return code.BadRequest, code.GetCodeMessage(code.BadRequest, "You don't have permission to delete this project!")
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

func (m *manager) Update(trx *gorm.DB, input *projectModel.Update) (int, any) {
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

	if *input.UpdateRole != "admin" {
		if *projectBase.CreatedBy != *input.UpdatedBy {
			if pmBase != nil {
				if *pmBase.ResourceUUID != *input.UpdateResUUID {
					return code.BadRequest, code.GetCodeMessage(code.BadRequest, "You don't have permission to update this project!")
				}
			} else {
				return code.BadRequest, code.GetCodeMessage(code.BadRequest, "You don't have permission to update this project!")
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
		// sync create project_resource
		for _, resource := range input.Resource {
			proRes := &projectResourceModel.Create{
				ProjectUUID:  *projectBase.ProjectUUID,
				ResourceUUID: resource.ResourceUUID,
				Role:         resource.Role,
				CreatedBy:    *input.UpdatedBy,
			}
			resourceList = append(resourceList, proRes)
			resourceUUIDs = append(resourceUUIDs, util.PointerString(resource.ResourceUUID))
		}

		_, err := m.ProjectResourceService.WithTrx(trx).CreateAll(resourceList)
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
