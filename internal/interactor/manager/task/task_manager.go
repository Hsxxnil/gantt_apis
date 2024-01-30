package task

import (
	"errors"
	"fmt"
	"github.com/bytedance/sonic"
	resourceManager "hta/internal/interactor/manager/resource"
	eventMarkModel "hta/internal/interactor/models/event_marks"
	projectResourceModel "hta/internal/interactor/models/project_resources"
	projectModel "hta/internal/interactor/models/projects"
	resourceModel "hta/internal/interactor/models/resources"
	taskResourceModel "hta/internal/interactor/models/task_resources"
	"hta/internal/interactor/pkg/util"
	eventMarkService "hta/internal/interactor/service/event_mark"
	projectService "hta/internal/interactor/service/project"
	projectResourceService "hta/internal/interactor/service/project_resource"
	resourceService "hta/internal/interactor/service/resource"
	taskResourceService "hta/internal/interactor/service/task_resource"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	taskModel "hta/internal/interactor/models/tasks"
	"hta/internal/interactor/pkg/util/code"
	"hta/internal/interactor/pkg/util/log"
	taskService "hta/internal/interactor/service/task"
)

type Manager interface {
	Create(trx *gorm.DB, input *taskModel.Create) (int, any)
	CreateAll(trx *gorm.DB, input []*taskModel.Create) (int, any)
	GetByProjectListNoPagination(input *taskModel.ProjectIDs) (int, any)
	GetByListNoPaginationNoSub(input *taskModel.Field) (int, any)
	GetBySingle(input *taskModel.Field) (int, any)
	Delete(trx *gorm.DB, input *taskModel.DeletedTaskUUIDs) (int, any)
	Update(trx *gorm.DB, input *taskModel.Update) (int, any)
	UpdateAll(trx *gorm.DB, input []*taskModel.Update) (int, any)
	Import(trx *gorm.DB, input *taskModel.Import) (int, any)
}

type manager struct {
	TaskService            taskService.Service
	ResourceService        resourceService.Service
	ResourceManager        resourceManager.Manager
	TaskResourceService    taskResourceService.Service
	ProjectService         projectService.Service
	ProjectResourceService projectResourceService.Service
	EventMarkService       eventMarkService.Service
}

func Init(db *gorm.DB) Manager {
	return &manager{
		TaskService:            taskService.Init(db),
		ResourceService:        resourceService.Init(db),
		TaskResourceService:    taskResourceService.Init(db),
		ResourceManager:        resourceManager.Init(db),
		ProjectService:         projectService.Init(db),
		ProjectResourceService: projectResourceService.Init(db),
		EventMarkService:       eventMarkService.Init(db),
	}
}

// getNextOutlineNumber is a helper function to generate the next outline number based on the last task's outline number.
func (m *manager) getNextOutlineNumber(isSubtask bool, outlineNumber *string, projectUUID string) (string, error) {
	lastTaskBase, err := m.TaskService.GetByLastOutlineNumber(&taskModel.Field{
		OutlineNumber: outlineNumber,
		ProjectUUID:   util.PointerString(projectUUID),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if isSubtask {
				log.Error(err)
				return *outlineNumber + ".1", nil
			}
			return "1", nil
		}

		log.Error(err)
		return "", err
	}

	newOutlineNumber, err := generateNewOutlineNumber(isSubtask, *lastTaskBase.OutlineNumber)
	if err != nil {
		log.Error(err)
		return "", err
	}
	return newOutlineNumber, nil
}

// createSubtasks is a helper function to create subtasks recursively for a parent task and returns the UUIDs of created subtasks.
func (m *manager) createSubtasks(trx *gorm.DB, parentTask *taskModel.Create, subtasks []*taskModel.Create, projectStart, projectEnd, minBaselineStart, maxBaselineEnd *time.Time) ([]*taskModel.Create, *time.Time, *time.Time, error) {
	var createList []*taskModel.Create
	for j, subBody := range subtasks {
		if subBody.BaselineStartDate != nil && subBody.BaselineEndDate != nil {
			// get the minimum baseline_start_date
			if minBaselineStart == nil || subBody.BaselineStartDate.Before(*minBaselineStart) {
				minBaselineStart = subBody.BaselineStartDate
			}

			// get the maximum baseline_end_date
			if maxBaselineEnd == nil || subBody.BaselineEndDate.After(*maxBaselineEnd) {
				maxBaselineEnd = subBody.BaselineEndDate
			}
		}

		newOutlineNumber := fmt.Sprintf("%s.%d", parentTask.OutlineNumber, j+1)
		subBody.OutlineNumber = newOutlineNumber
		subBody.IsSubTask = true
		subBody.CreatedBy = parentTask.CreatedBy
		subBody.ProjectUUID = parentTask.ProjectUUID
		// transform segments from struct array to string
		if len(subBody.Segments) > 0 {
			// check if it is the most sub-task
			if len(subBody.Subtask) > 0 {
				return nil, nil, nil, errors.New("the parent task cannot be segmented")
			}
			segJson, err := sonic.Marshal(subBody.Segments)
			if err != nil {
				log.Error(err)
				return nil, nil, nil, err
			}
			subBody.Segment = string(segJson)
		}

		// transform indicators from struct array to string
		if len(subBody.Indicators) > 0 {
			indJson, err := sonic.Marshal(subBody.Indicators)
			if err != nil {
				log.Error(err)
				return nil, nil, nil, err
			}
			subBody.Indicator = string(indJson)
		}
		createList = append(createList, subBody)

		// handle possible subtasks
		if len(subBody.Subtask) > 0 {
			subSubtasks, subMinBaselineStart, subMaxBaselineEnd, err := m.createSubtasks(trx, subBody, subBody.Subtask, projectStart, projectEnd, minBaselineStart, maxBaselineEnd)
			if err != nil {
				return nil, nil, nil, err
			}
			createList = append(createList, subSubtasks...)

			// compare the minimum baseline_start_date and maximum baseline_end_date of the subtasks
			if subBody.BaselineStartDate != nil && subBody.BaselineEndDate != nil {
				if minBaselineStart == nil || subBody.BaselineStartDate.Before(*subMinBaselineStart) {
					minBaselineStart = subBody.BaselineStartDate
				}
				if maxBaselineEnd == nil || subBody.BaselineEndDate.After(*subMaxBaselineEnd) {
					maxBaselineEnd = subBody.BaselineEndDate
				}
			}
		}
	}

	return createList, minBaselineStart, maxBaselineEnd, nil
}

// updateSubtasks is a helper function to update subtasks recursively for a parent task and returns the UUIDs of updated subtasks.
func (m *manager) updateSubtasks(trx *gorm.DB, parentTask *taskModel.Update, subtasks []*taskModel.Update, minBaselineStart, maxBaselineEnd *time.Time, role, resUUID *string, taskMap map[string]*taskModel.Single) ([]*string, []*taskModel.Update, []map[string][]*resourceModel.TaskSingle, *time.Time, *time.Time, error) {
	var (
		taskResMapList []map[string][]*resourceModel.TaskSingle
		updateList     []*taskModel.Update
		TaskUUIDs      []*string
	)

	for j, subBody := range subtasks {
		// check if the update_by is the task's creator or assigned resource
		if *role != "admin" && resUUID != nil && taskMap[subBody.TaskUUID] != nil {
			if taskMap[subBody.TaskUUID].CreatedBy != *subBody.UpdatedBy {
				assigned := false
				for _, res := range taskMap[subBody.TaskUUID].Resources {
					if res.ResourceUUID == *resUUID {
						assigned = true
						break
					}
				}
				if !assigned {
					continue
				}
			}
		}

		if subBody.BaselineStartDate != nil && subBody.BaselineEndDate != nil {
			// get the minimum baseline_start_date
			if minBaselineStart == nil || subBody.BaselineStartDate.Before(*minBaselineStart) {
				minBaselineStart = subBody.BaselineStartDate
			}

			// get the maximum baseline_end_date
			if maxBaselineEnd == nil || subBody.BaselineEndDate.After(*maxBaselineEnd) {
				maxBaselineEnd = subBody.BaselineEndDate
			}
		}

		subBody.UpdatedBy = parentTask.UpdatedBy
		subBody.ProjectUUID = parentTask.ProjectUUID
		subBody.IsSubTask = util.PointerBool(true)
		// update task's outline_number
		newOutlineNumber := fmt.Sprintf("%s.%d", *parentTask.OutlineNumber, j+1)
		subBody.OutlineNumber = util.PointerString(newOutlineNumber)

		// transform segments from struct array to string
		if len(subBody.Segments) > 0 {
			// check if it is the most sub-task
			if len(subBody.Subtask) > 0 {
				return nil, nil, nil, nil, nil, errors.New("the parent task cannot be segmented")
			}
			segJson, err := sonic.Marshal(subBody.Segments)
			if err != nil {
				log.Error(err)
				return nil, nil, nil, nil, nil, err
			}
			subBody.Segment = util.PointerString(string(segJson))
		} else {
			subBody.Segment = nil
		}

		// transform indicators from struct array to string
		if len(subBody.Indicators) > 0 {
			indJson, err := sonic.Marshal(subBody.Indicators)
			if err != nil {
				log.Error(err)
				return nil, nil, nil, nil, nil, err
			}
			subBody.Indicator = util.PointerString(string(indJson))
		} else {
			subBody.Indicator = nil
		}

		// sync update task_resource
		if len(subBody.Resources) > 0 {
			TaskUUIDs = append(TaskUUIDs, util.PointerString(subBody.TaskUUID))
			taskResources := make(map[string][]*resourceModel.TaskSingle)
			taskResources[subBody.TaskUUID] = subBody.Resources
			taskResMapList = append(taskResMapList, taskResources)
		}

		// update subtask
		updateList = append(updateList, subBody)

		// handle possible subtasks
		if len(subBody.Subtask) > 0 {
			subTaskUUIDs, subUpdateList, subTaskResMapList, subMinBaselineStart, subMaxBaselineEnd, err := m.updateSubtasks(trx, subBody, subBody.Subtask, minBaselineStart, maxBaselineEnd, role, resUUID, taskMap)
			if err != nil {
				return nil, nil, nil, nil, nil, err
			}
			taskResMapList = append(taskResMapList, subTaskResMapList...)
			updateList = append(updateList, subUpdateList...)
			TaskUUIDs = append(TaskUUIDs, subTaskUUIDs...)

			if subBody.BaselineStartDate != nil && subBody.BaselineEndDate != nil {
				// compare the minimum baseline_start_date and maximum baseline_end_date of the subtasks
				if minBaselineStart == nil || subBody.BaselineStartDate.Before(*subMinBaselineStart) {
					minBaselineStart = subBody.BaselineStartDate
				}
				if maxBaselineEnd == nil || subBody.BaselineEndDate.After(*subMaxBaselineEnd) {
					maxBaselineEnd = subBody.BaselineEndDate
				}
			}
		}
	}

	return TaskUUIDs, updateList, taskResMapList, minBaselineStart, maxBaselineEnd, nil
}

// syncCreateTaskResources is a helper function to synchronize the creation of task_resource associations for tasks of a project.
func (m *manager) syncCreateTaskResources(trx *gorm.DB, taskResources []map[string][]*resourceModel.TaskSingle, createdBy string, projectID string, proResMap map[string]*projectResourceModel.Single) error {
	var (
		resList    []*taskResourceModel.Create
		proResList []*projectResourceModel.Create
	)

	for _, taskRes := range taskResources {
		for taskUUID, resources := range taskRes {
			for _, res := range resources {
				resList = append(resList, &taskResourceModel.Create{
					TaskUUID:     taskUUID,
					ResourceUUID: res.ResourceUUID,
					Unit:         res.Unit,
					CreatedBy:    createdBy,
				})

				if proResMap[res.ResourceUUID] == nil {
					proResList = append(proResList, &projectResourceModel.Create{
						ProjectUUID:  projectID,
						ResourceUUID: res.ResourceUUID,
						IsEditable:   true,
						CreatedBy:    createdBy,
					})
				}

			}
		}
	}

	_, err := m.TaskResourceService.WithTrx(trx).CreateAll(resList)
	if err != nil {
		return err
	}

	// sync create project_resource
	if len(proResList) > 0 {
		_, err = m.ProjectResourceService.WithTrx(trx).CreateAll(proResList)
		if err != nil {
			return err
		}
	}

	return nil
}

// syncDeleteTaskResources is a helper function to synchronize the deletion of task_resource associations for a tasks of a project.
func (m *manager) syncDeleteTaskResources(trx *gorm.DB, TaskUUID *string, TaskUUIDs []*string, isBatch bool) error {
	if isBatch && len(TaskUUIDs) > 0 {
		err := m.TaskResourceService.WithTrx(trx).Delete(&taskResourceModel.Field{
			TaskUUIDs: TaskUUIDs,
		})
		if err != nil {
			return err
		}

	} else {
		err := m.TaskResourceService.WithTrx(trx).Delete(&taskResourceModel.Field{
			TaskUUID: TaskUUID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// syncUpdateProjectStartEndDate is a helper function to synchronize the update project start and end dates.
func (m *manager) syncUpdateProjectStartEndDate(trx *gorm.DB, projectID *string, TaskUUIDs []*string, start, end *time.Time) error {
	if projectID == nil {
		return errors.New("ProjectUUID is null")
	}
	var minBaselineStart, maxBaselineEnd *time.Time
	// get the minimum baseline_start_date and maximum baseline_end_date of the tasks
	taskBase, err := m.TaskService.GetByMinStartMaxEnd(&taskModel.Field{
		ProjectUUID:      projectID,
		DeletedTaskUUIDs: TaskUUIDs,
	})
	if err != nil {
		return err
	}

	if len(taskBase) > 0 {
		// find the minimum baseline_start_date and maximum baseline_end_date in taskBase
		for _, task := range taskBase {
			if task.BaselineStartDate != nil && task.BaselineEndDate != nil {
				if minBaselineStart == nil || task.BaselineStartDate.Before(*minBaselineStart) {
					minBaselineStart = task.BaselineStartDate
				}
				if maxBaselineEnd == nil || task.BaselineEndDate.After(*maxBaselineEnd) {
					maxBaselineEnd = task.BaselineEndDate
				}
			}
		}
	}

	// compare the minimum baseline_start_date and maximum baseline_end_date of the tasks
	if start != nil && end != nil {
		if minBaselineStart == nil || start.Before(*minBaselineStart) {
			minBaselineStart = start
		}
		if maxBaselineEnd == nil || end.After(*maxBaselineEnd) {
			maxBaselineEnd = end
		}
	}

	// sync update project's start and end dates
	err = m.ProjectService.WithTrx(trx).Update(&projectModel.Update{
		ProjectUUID: *projectID,
		StartDate:   minBaselineStart,
		EndDate:     maxBaselineEnd,
	})
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

// generateNewOutlineNumber is a helper function used to generate a new outline number within the "getNextOutlineNumber" function.
func generateNewOutlineNumber(isSubtask bool, lastOutlineNumber string) (string, error) {
	var newOutlineNumber string
	outlines := strings.Split(lastOutlineNumber, ".")

	if isSubtask {
		lastTaskOutline, err := strconv.Atoi(outlines[len(outlines)-1])
		if err != nil {
			log.Error(err)
			return "", err
		}
		newOutlineNumber = fmt.Sprintf("%s.%d", strings.Join(outlines[:len(outlines)-1], "."), lastTaskOutline+1)

	} else {
		lastTaskOutline, err := strconv.Atoi(outlines[0])
		if err != nil {
			log.Error(err)
			return "", err
		}
		newOutlineNumber = strconv.Itoa(lastTaskOutline + 1)
	}
	return newOutlineNumber, nil
}

// getParentOutlineNumber is a helper function to extract the parent outline number.
func getParentOutlineNumber(outlineNumber string) string {
	parts := strings.Split(outlineNumber, ".")
	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], ".")
	}
	return ""
}

// findParentTaskByOutlineNumber is a helper function to find the parent task by outline number.
func findParentTaskByOutlineNumber(tasks []*taskModel.Single, outlineNumber string) *taskModel.Single {
	for _, task := range tasks {
		if task.OutlineNumber == outlineNumber {
			return task
		}
	}
	return nil
}

// assembleToCreateAll is a helper function to recursively assemble a hierarchical structure of tasks.
func assembleToCreateAll(tasks []*taskModel.Create, taskRecordIdx map[string]int, task *taskModel.Create, outlineNumber, selectOutlineNumber, projectID string, resUUID, role *string) []*taskModel.Create {
	task.ProjectUUID = projectID
	outlineNumberSplit := strings.Split(outlineNumber, ".")

	// handle the main task
	if len(outlineNumberSplit) == 1 {
		// record the outline_number of the current main task
		if outlineNumber != "" {
			selectOutlineNumber += "." + outlineNumber
		}
		// map the outline_number of the current main task to the index position of allTask
		taskRecordIdx[selectOutlineNumber] = len(tasks)
		tasks = append(tasks, task)
	} else {
		// handle the current subtask
		var newOutlineNumber string
		for i := 1; i < len(outlineNumberSplit); i++ {
			newOutlineNumber += outlineNumberSplit[i]
			// if it is not the last subtask, add a separator to the outline_number
			if i != len(outlineNumberSplit)-1 {
				newOutlineNumber += "."
			}
		}

		// append the outline_number of the current subtask to the outline_number of its parent task
		if selectOutlineNumber == "" {
			selectOutlineNumber = outlineNumberSplit[0]
		} else {
			selectOutlineNumber += "." + outlineNumberSplit[0]
		}

		// handle possible subtasks
		subTasks := assembleToCreateAll(tasks[taskRecordIdx[selectOutlineNumber]].Subtask, taskRecordIdx, task, newOutlineNumber, selectOutlineNumber, projectID, resUUID, role)
		tasks[taskRecordIdx[selectOutlineNumber]].Subtask = subTasks
	}

	for _, task := range tasks {
		if resUUID != nil {
			task.ResUUID = resUUID
		}
		if role != nil {
			task.Role = role
		}
	}

	return tasks
}

func (m *manager) Create(trx *gorm.DB, input *taskModel.Create) (int, any) {
	defer trx.Rollback()

	// get the project's info
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

	// check if the user is a project member
	proResBase, err := m.ProjectResourceService.GetBySingle(&projectResourceModel.Field{
		ProjectUUID:  util.PointerString(input.ProjectUUID),
		ResourceUUID: input.ResUUID,
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// if the user is a project member, check if the user can edit or delete tasks of the project
	if proResBase != nil {
		if !*proResBase.IsEditable {
			log.Error("You are not allowed to create tasks for this project.")
			return code.BadRequest, code.GetCodeMessage(code.BadRequest, "You are not allowed to create tasks for this project.")
		}
	} else {
		// if the user is not a project member, check if the user is an admin or the creator of the project
		if *input.Role != "admin" {
			if *projectBase.CreatedBy != input.CreatedBy {
				log.Error("You are not allowed to create tasks for this project.")
				return code.BadRequest, code.GetCodeMessage(code.BadRequest, "You are not allowed to create tasks for this project.")
			}
		}
	}

	// determine if the task is a subtask
	if input.ParentUUID != nil {
		parentBase, err := m.TaskService.GetBySingle(&taskModel.Field{
			TaskUUID: *input.ParentUUID,
		})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
			}

			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
		input.ProjectUUID = *parentBase.ProjectUUID

		// get the last outline_number of the same level
		input.OutlineNumber, err = m.getNextOutlineNumber(true, parentBase.OutlineNumber, input.ProjectUUID)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
		input.IsSubTask = true

		// check if it is the most sub-task
		if parentBase.Segment != nil {
			if *parentBase.Segment != "" {
				log.Info("Please remove the task segmentation first.")
				return code.BadRequest, code.GetCodeMessage(code.BadRequest, "Please remove the task segmentation first.")
			}
		}

		// transform segments from struct array to string
		if len(input.Segments) > 0 {
			segJson, _ := sonic.Marshal(input.Segments)
			input.Segment = string(segJson)
		}

		// transform indicators from struct array to string
		if len(input.Indicators) > 0 {
			indJson, _ := sonic.Marshal(input.Indicators)
			input.Indicator = string(indJson)
		}

	} else {
		// the main task with subtasks cannot be segmented
		if len(input.Segments) > 0 {
			if len(input.Subtask) > 0 {
				log.Info("The parent task cannot be segmented.")
				return code.BadRequest, code.GetCodeMessage(code.BadRequest, "The parent task cannot be segmented.")
			}

			// transform segments from struct array to string
			segJson, _ := sonic.Marshal(input.Segments)
			input.Segment = string(segJson)
		}

		// transform indicators from struct array to string
		if len(input.Indicators) > 0 {
			indJson, _ := sonic.Marshal(input.Indicators)
			input.Indicator = string(indJson)
		}

		// get the new outline_number
		newOutlineNumber, err := m.getNextOutlineNumber(false, nil, input.ProjectUUID)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
		input.OutlineNumber = newOutlineNumber
	}

	taskBase, err := m.TaskService.WithTrx(trx).Create(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	if len(input.Resources) > 0 {
		// get all resources of the project
		proResBase, err := m.ProjectResourceService.GetByListNoPagination(&projectResourceModel.Field{
			ProjectUUID: util.PointerString(input.ProjectUUID),
		})
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
		}

		var proRes []*projectResourceModel.Single
		proResByte, err := sonic.Marshal(proResBase)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}

		err = sonic.Unmarshal(proResByte, &proRes)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}

		// create a map of resourceUUID
		proResMap := make(map[string]*projectResourceModel.Single)
		if len(proRes) > 0 {
			for _, res := range proRes {
				proResMap[res.ResourceUUID] = res
			}
		}

		// sync create task_resource
		taskResMapList := []map[string][]*resourceModel.TaskSingle{
			{*taskBase.TaskUUID: input.Resources},
		}
		err = m.syncCreateTaskResources(trx, taskResMapList, input.CreatedBy, input.ProjectUUID, proResMap)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// sync update project's start and end dates
	err = m.syncUpdateProjectStartEndDate(trx, util.PointerString(input.ProjectUUID), nil, input.BaselineStartDate, input.BaselineEndDate)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, taskBase.TaskUUID)
}

func (m *manager) CreateAll(trx *gorm.DB, input []*taskModel.Create) (int, any) {
	defer trx.Rollback()

	var (
		createList                       []*taskModel.Create
		taskResMapList                   []map[string][]*resourceModel.TaskSingle
		minBaselineStart, maxBaselineEnd *time.Time
	)

	// get all resources of the project
	proResBase, err := m.ProjectResourceService.GetByListNoPagination(&projectResourceModel.Field{
		ProjectUUID: util.PointerString(input[0].ProjectUUID),
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	var proRes []*projectResourceModel.Single
	proResByte, err := sonic.Marshal(proResBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(proResByte, &proRes)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// create a map of resourceUUID
	proResMap := make(map[string]*projectResourceModel.Single)
	if len(proRes) > 0 {
		for _, res := range proRes {
			proResMap[res.ResourceUUID] = res
		}
	}

	// get the new outline_number
	topOutlineNumber, err := m.getNextOutlineNumber(false, nil, input[0].ProjectUUID)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	newOutlineNumber, err := strconv.Atoi(topOutlineNumber)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// get the project's info
	projectBase, err := m.ProjectService.GetBySingle(&projectModel.Field{
		ProjectUUID: input[0].ProjectUUID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// if the user is a project member, check if the user can edit or delete tasks of the project
	if proResMap[*input[0].ResUUID] != nil {
		if !proResMap[*input[0].ResUUID].IsEditable {
			log.Error("You are not allowed to create tasks for this project.")
			return code.BadRequest, code.GetCodeMessage(code.BadRequest, "You are not allowed to create tasks for this project.")
		}
	} else {
		// if the user is not a project member, check if the user is an admin or the creator of the project
		if *input[0].Role != "admin" {
			if *projectBase.CreatedBy != input[0].CreatedBy {
				log.Error("You are not allowed to create tasks for this project.")
				return code.BadRequest, code.GetCodeMessage(code.BadRequest, "You are not allowed to create tasks for this project.")
			}
		}
	}

	// create the main task
	for i, inputBody := range input {
		if inputBody.BaselineStartDate != nil && inputBody.BaselineEndDate != nil {
			// get the minimum baseline_start_date
			if minBaselineStart == nil || inputBody.BaselineStartDate.Before(*minBaselineStart) {
				minBaselineStart = inputBody.BaselineStartDate
			}

			// get the maximum baseline_end_date
			if maxBaselineEnd == nil || inputBody.BaselineEndDate.After(*maxBaselineEnd) {
				maxBaselineEnd = inputBody.BaselineEndDate
			}
		}

		// the main task with subtasks cannot be segmented
		if len(inputBody.Segments) > 0 {
			if len(inputBody.Subtask) > 0 {
				log.Info("The parent task cannot be segmented.")
				return code.BadRequest, code.GetCodeMessage(code.BadRequest, "The parent task cannot be segmented.")
			}

			// transform segments from struct array to string
			segJson, _ := sonic.Marshal(inputBody.Segments)
			inputBody.Segment = string(segJson)
		}

		// transform indicators from struct array to string
		if len(inputBody.Indicators) > 0 {
			indJson, _ := sonic.Marshal(inputBody.Indicators)
			inputBody.Indicator = string(indJson)
		}

		// get the new outline_number
		inputBody.OutlineNumber = strconv.Itoa(newOutlineNumber + i)
		createList = append(createList, inputBody)

		// create subtasks
		if len(inputBody.Subtask) > 0 {
			subSubtasks, subMinBaselineStart, subMaxBaselineEnd, err := m.createSubtasks(trx, inputBody, inputBody.Subtask, projectBase.StartDate, projectBase.EndDate, minBaselineStart, maxBaselineEnd)
			if err != nil {
				if err.Error() == "the parent task cannot be segmented" {
					return code.BadRequest, code.GetCodeMessage(code.BadRequest, "The parent task cannot be segmented.")
				}

				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
			createList = append(createList, subSubtasks...)

			// compare the minimum baseline_start_date and maximum baseline_end_date of the subtasks
			if inputBody.BaselineStartDate != nil && inputBody.BaselineEndDate != nil {
				if minBaselineStart == nil || inputBody.BaselineStartDate.Before(*subMinBaselineStart) {
					minBaselineStart = inputBody.BaselineStartDate
				}
				if maxBaselineEnd == nil || inputBody.BaselineEndDate.After(*subMaxBaselineEnd) {
					maxBaselineEnd = inputBody.BaselineEndDate
				}
			}
		}
	}

	tasksBase, err := m.TaskService.WithTrx(trx).CreateAll(createList)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// sync create task_resource
	for i, taskBase := range tasksBase {
		if len(createList[i].Resources) > 0 {
			taskResources := make(map[string][]*resourceModel.TaskSingle)
			taskResources[*taskBase.TaskUUID] = createList[i].Resources
			taskResMapList = append(taskResMapList, taskResources)
		}
	}

	// sync create task_resource
	if len(taskResMapList) > 0 {
		err = m.syncCreateTaskResources(trx, taskResMapList, createList[0].CreatedBy, createList[0].ProjectUUID, proResMap)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// sync update project's start and end dates
	err = m.syncUpdateProjectStartEndDate(trx, util.PointerString(input[0].ProjectUUID), nil, minBaselineStart, maxBaselineEnd)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, "Successful create!")
}

func (m *manager) GetByProjectListNoPagination(input *taskModel.ProjectIDs) (int, any) {
	output := &taskModel.List{}

	// get projects
	projectBase, err := m.ProjectService.GetByListNoPagination(&projectModel.Field{
		ProjectUUIDs: input.Projects,
	})
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	if len(projectBase) == 0 {
		log.Error("record not found")
		return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, "record not found")
	}

	var projects []*projectModel.Single
	projectByte, err := sonic.Marshal(projectBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(projectByte, &projects)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// create a map of projectUUID
	projectMap := make(map[string]*projectModel.Single)
	for _, project := range projects {
		projectMap[project.ProjectUUID] = project
	}

	// check the user's permission of the project
	isPM := false
	isEditable := false
	if len(input.Projects) == 1 {
		if *input.Role != "admin" {
			// check the user can edit or delete tasks of the project
			proResBase, err := m.ProjectResourceService.GetBySingle(&projectResourceModel.Field{
				ProjectUUID:  input.Projects[0],
				ResourceUUID: input.ResUUID,
			})
			if err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					log.Error(err)
					return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
				}
			}

			if proResBase != nil {
				// check if the user has permission to edit or delete the project
				if *proResBase.IsEditable {
					isEditable = true
				}

				// check if the user is a project manager
				if *proResBase.Role == "PM" {
					isPM = true
				}
			}
		}
	}

	// get project tasks
	projectTaskMap := make(map[*string][]*taskModel.Single)
	var (
		wg               sync.WaitGroup
		mu               sync.Mutex
		projectStartDate *time.Time
	)

	// make an error channel
	goroutineErr := make(chan error)
	// create goroutine
	for _, projectsUUID := range input.Projects {
		wg.Add(1)
		go func(projectsUUID *string) {
			defer wg.Done()
			mu.Lock()
			defer mu.Unlock()

			// get tasks
			taskBase, err := m.TaskService.GetByListNoPagination(&taskModel.Field{
				ProjectUUID: projectsUUID,
				Filter: taskModel.Filter{
					FilterMilestone: input.FilterMilestone,
				},
			})
			if err != nil {
				log.Error(err)
				goroutineErr <- err
			}

			var projectTasks []*taskModel.Single
			taskByte, err := sonic.Marshal(taskBase)
			if err != nil {
				log.Error(err)
				goroutineErr <- err
			}

			err = sonic.Unmarshal(taskByte, &projectTasks)
			if err != nil {
				log.Error(err)
				goroutineErr <- err
			}

			// create a map of taskUUID
			taskMap := make(map[string]*taskModel.Single)
			for _, task := range projectTasks {
				taskMap[task.TaskUUID] = task
			}

			for i, task := range projectTasks {
				task.CreatedBy = *taskBase[i].CreatedByUsers.Name
				task.UpdatedBy = *taskBase[i].UpdatedByUsers.Name
				for j, file := range taskBase[i].S3Files {
					task.Files[j].CreatedBy = *file.CreatedByUsers.Name
				}

				// transform segments to array
				var segments []taskModel.Segments
				err = util.DecodeJSONToSlice(*taskBase[i].Segment, &segments)
				if err != nil {
					log.Error(err)
					goroutineErr <- err
				}
				task.Segments = segments

				// transform indicator to array
				var indicators []taskModel.Indicators
				err = util.DecodeJSONToSlice(*taskBase[i].Indicator, &indicators)
				if err != nil {
					log.Error(err)
					goroutineErr <- err
				}
				task.Indicators = indicators
				if len(indicators) > 0 {
					task.IndicatorsName = indicators[0].Name
					task.IndicatorsIconClass = indicators[0].IconClass
					task.IndicatorsToolTip = indicators[0].ToolTip
				}

				for j, res := range taskBase[i].TaskResources {
					task.Resources[j].ResourceUUID = *res.Resources.ResourceUUID
					task.Resources[j].ResourceID = *res.Resources.Resources.ResourceID
					task.Resources[j].ResourceName = *res.Resources.Resources.ResourceName
					task.Resources[j].Email = *res.Resources.Resources.Email
					task.Resources[j].Phone = *res.Resources.Resources.Phone
					task.Resources[j].StandardCost = *res.Resources.Resources.StandardCost
					task.Resources[j].TotalCost = *res.Resources.Resources.TotalCost
					task.Resources[j].TotalLoad = *res.Resources.Resources.TotalLoad
					task.Resources[j].IsExpand = *res.Resources.Resources.IsExpand
					task.Resources[j].Role = *res.Resources.Role

					// transform resource_groups to array
					var resourceGroup []string
					err = sonic.Unmarshal([]byte(*res.Resources.Resources.ResourceGroup), &resourceGroup)
					if err != nil {
						log.Error(err)
						goroutineErr <- err
					}
					task.Resources[j].ResourceGroups = resourceGroup
				}

				// check the user can edit or delete the task
				if *input.Role == "admin" || isPM {
					task.IsEditable = true
				} else {
					if *taskBase[i].CreatedBy == *input.UserID {
						task.IsEditable = true
					} else {
						for _, res := range task.Resources {
							if res.ResourceUUID == *input.ResUUID {
								task.IsEditable = true
								break
							}
						}
					}
				}

				if !input.FilterMilestone {
					// determine if the task is a subtask
					if task.IsSubTask {
						// get the parent outline number
						parentOutlineNumber := getParentOutlineNumber(*taskBase[i].OutlineNumber)
						if parentOutlineNumber != "" {
							// find tasks with the same parent outline number
							parentTask := findParentTaskByOutlineNumber(projectTasks, parentOutlineNumber)
							if parentTask != nil {
								parentTask.Subtask = append(parentTask.Subtask, taskMap[task.TaskUUID])
							}
						}
					}
				}
			}

			projectTaskMap[projectsUUID] = projectTasks
		}(projectsUUID)
	}

	// wait until all goroutines are finished
	wg.Wait()
	close(goroutineErr)

	// check if goroutine has error
	err = <-goroutineErr
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	for _, projectsUUID := range input.Projects {
		if len(projectTaskMap[projectsUUID]) > 0 {
			if len(input.Projects) == 1 {
				// if the user is admin or project manager, check if the user can edit or delete tasks of the project
				if *input.Role == "admin" || isPM || isEditable {
					output.IsEditable = true
				} else {
					// if the user is not a project member, check if the user is the creator of the project
					if projectMap[*projectsUUID].CreatedBy == *input.UserID {
						output.IsEditable = true
					}
				}

				if !input.FilterMilestone {
					// filter out the subtasks that have been included in each SubTask
					var filteredTasks []*taskModel.Single
					for _, task := range projectTaskMap[projectsUUID] {
						if !task.IsSubTask {
							filteredTasks = append(filteredTasks, task)
						}
					}
					output.Tasks = filteredTasks
				} else {
					output.Tasks = projectTaskMap[projectsUUID]
				}
				// return project status
				output.ProjectStatus = *projectBase[0].Status

				// get project start date
				for _, task := range output.Tasks {
					if task.BaselineStartDate != nil {
						if projectStartDate == nil || task.BaselineStartDate.Before(*projectStartDate) {
							projectStartDate = task.BaselineStartDate
						}
					}

					if task.StartDate != nil {
						if projectStartDate == nil || task.StartDate.Before(*projectStartDate) {
							projectStartDate = task.StartDate
						}
					}

					if task.Subtask != nil {
						for _, subtask := range task.Subtask {
							if subtask.BaselineStartDate != nil {
								if projectStartDate == nil || subtask.BaselineStartDate.Before(*projectStartDate) {
									projectStartDate = subtask.BaselineStartDate
								}
							}

							if subtask.StartDate != nil {
								if projectStartDate == nil || subtask.StartDate.Before(*projectStartDate) {
									projectStartDate = subtask.StartDate
								}
							}
						}
					}
				}
				output.ProjectStartDate = projectStartDate

			} else {
				projectTask := &taskModel.Single{
					TaskUUID:  projectMap[*projectsUUID].ProjectUUID,
					TaskID:    projectMap[*projectsUUID].ProjectID,
					TaskName:  projectMap[*projectsUUID].ProjectName,
					IsSubTask: false,
				}

				var projectSubtasks []*taskModel.Single
				for _, task := range projectTaskMap[projectsUUID] {
					projectSubtasks = append(projectSubtasks, task)
				}
				projectTask.Subtask = projectSubtasks

				if !input.FilterMilestone {
					// filter out the subtasks that have been included in each SubTask
					var filteredTasks []*taskModel.Single
					for _, task := range projectTask.Subtask {
						task.Predecessor = ""
						if !task.IsSubTask {
							filteredTasks = append(filteredTasks, task)
						}
					}
					projectTask.Subtask = filteredTasks
				} else {
					projectTask.Subtask = projectSubtasks
				}

				// add project to output.Tasks
				output.Tasks = append(output.Tasks, projectTask)

				// get project start date
				for _, task := range output.Tasks {
					if task.BaselineStartDate != nil {
						if projectStartDate == nil || task.BaselineStartDate.Before(*projectStartDate) {
							projectStartDate = task.BaselineStartDate
						}
					}

					if task.StartDate != nil {
						if projectStartDate == nil || task.StartDate.Before(*projectStartDate) {
							projectStartDate = task.StartDate
						}
					}

					if task.Subtask != nil {
						for _, subtask := range task.Subtask {
							if subtask.BaselineStartDate != nil {
								if projectStartDate == nil || subtask.BaselineStartDate.Before(*projectStartDate) {
									projectStartDate = subtask.BaselineStartDate
								}
							}

							if subtask.StartDate != nil {
								if projectStartDate == nil || subtask.StartDate.Before(*projectStartDate) {
									projectStartDate = subtask.StartDate
								}
							}
						}
					}
				}
				output.ProjectStartDate = projectStartDate
			}
		} else {
			if len(input.Projects) == 1 {
				// return project status
				output.ProjectStatus = *projectBase[0].Status

				// if the user is admin or project manager, check if the user can edit or delete tasks of the project
				if *input.Role == "admin" || isPM || isEditable {
					output.IsEditable = true
				} else {
					// if the user is not a project member, check if the user is the creator of the project
					if projectMap[*projectsUUID].CreatedBy == *input.UserID {
						output.IsEditable = true
					}
				}
			}
		}

		// check if output.Tasks is null, return empty array
		if output.Tasks == nil {
			output.Tasks = []*taskModel.Single{}
		}

		// get event_marks
		var eventMarks []*eventMarkModel.Single
		eventMarkBase, err := m.EventMarkService.GetByListNoPagination(&eventMarkModel.Field{
			ProjectUUID: projectsUUID,
		})
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
		eventMarkByte, err := sonic.Marshal(eventMarkBase)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}

		err = sonic.Unmarshal(eventMarkByte, &eventMarks)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}

		// add event_marks to output.Tasks
		output.EventMarks = append(output.EventMarks, eventMarks...)
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetByListNoPaginationNoSub(input *taskModel.Field) (int, any) {
	output := &taskModel.List{}
	taskBase, err := m.TaskService.GetByListNoPagination(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	taskByte, err := sonic.Marshal(taskBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(taskByte, &output.Tasks)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	for i, task := range output.Tasks {
		task.CreatedBy = *taskBase[i].CreatedByUsers.Name
		task.UpdatedBy = *taskBase[i].UpdatedByUsers.Name
		for j, file := range taskBase[i].S3Files {
			task.Files[j].CreatedBy = *file.CreatedByUsers.Name
		}

		// transform segments to array
		var segments []taskModel.Segments
		err = util.DecodeJSONToSlice(*taskBase[i].Segment, &segments)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
		task.Segments = segments

		// transform indicator to array
		var indicators []taskModel.Indicators
		err = util.DecodeJSONToSlice(*taskBase[i].Indicator, &indicators)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
		task.Indicators = indicators
		if len(indicators) > 0 {
			task.IndicatorsName = indicators[0].Name
			task.IndicatorsIconClass = indicators[0].IconClass
			task.IndicatorsToolTip = indicators[0].ToolTip
		}

		for j, res := range taskBase[i].TaskResources {
			task.Resources[j].ResourceUUID = *res.Resources.ResourceUUID
			task.Resources[j].ResourceID = *res.Resources.Resources.ResourceID
			task.Resources[j].ResourceName = *res.Resources.Resources.ResourceName
			task.Resources[j].Email = *res.Resources.Resources.Email
			task.Resources[j].Phone = *res.Resources.Resources.Phone
			task.Resources[j].StandardCost = *res.Resources.Resources.StandardCost
			task.Resources[j].TotalCost = *res.Resources.Resources.TotalCost
			task.Resources[j].TotalLoad = *res.Resources.Resources.TotalLoad
			task.Resources[j].IsExpand = *res.Resources.Resources.IsExpand
			task.Resources[j].Role = *res.Resources.Role

			// transform resource_groups to array
			var resourceGroup []string
			err = sonic.Unmarshal([]byte(*res.Resources.Resources.ResourceGroup), &resourceGroup)
			if err != nil {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
			task.Resources[j].ResourceGroups = resourceGroup
		}
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetBySingle(input *taskModel.Field) (int, any) {
	taskBase, err := m.TaskService.GetBySingle(input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output := &taskModel.Single{}
	taskByte, _ := sonic.Marshal(taskBase)
	err = sonic.Unmarshal(taskByte, &output)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output.CreatedBy = *taskBase.CreatedByUsers.Name
	output.UpdatedBy = *taskBase.UpdatedByUsers.Name
	for j, file := range taskBase.S3Files {
		output.Files[j].CreatedBy = *file.CreatedByUsers.Name
	}

	// transform segments to array
	var segments []taskModel.Segments
	err = util.DecodeJSONToSlice(*taskBase.Segment, &segments)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	output.Segments = segments

	// transform indicator to array
	var indicators []taskModel.Indicators
	err = util.DecodeJSONToSlice(*taskBase.Indicator, &indicators)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	output.Indicators = indicators
	if len(indicators) > 0 {
		output.IndicatorsName = indicators[0].Name
		output.IndicatorsIconClass = indicators[0].IconClass
		output.IndicatorsToolTip = indicators[0].ToolTip
	}

	for i, res := range taskBase.TaskResources {
		output.Resources[i].ResourceUUID = *res.Resources.ResourceUUID
		output.Resources[i].ResourceID = *res.Resources.Resources.ResourceID
		output.Resources[i].ResourceName = *res.Resources.Resources.ResourceName
		output.Resources[i].Email = *res.Resources.Resources.Email
		output.Resources[i].Phone = *res.Resources.Resources.Phone
		output.Resources[i].StandardCost = *res.Resources.Resources.StandardCost
		output.Resources[i].TotalCost = *res.Resources.Resources.TotalCost
		output.Resources[i].TotalLoad = *res.Resources.Resources.TotalLoad
		output.Resources[i].IsExpand = *res.Resources.Resources.IsExpand
		output.Resources[i].Role = *res.Resources.Role

		// transform resource_groups to array
		var resourceGroup []string
		err = sonic.Unmarshal([]byte(*res.Resources.Resources.ResourceGroup), &resourceGroup)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
		output.Resources[i].ResourceGroups = resourceGroup
	}
	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) Delete(trx *gorm.DB, input *taskModel.DeletedTaskUUIDs) (int, any) {
	defer trx.Rollback()

	// check the update_by has the permission to delete the project's tasks
	if *input.Role != "admin" {
		// search project_resource
		proResBase, err := m.ProjectResourceService.GetBySingle(&projectResourceModel.Field{
			ProjectUUID:  input.ProjectUUID,
			ResourceUUID: input.ResUUID,
		})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
			}

			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}

		if proResBase != nil {
			if !*proResBase.IsEditable {
				log.Info("The user don't have permission to update the project's tasks.")
				return code.BadRequest, code.GetCodeMessage(code.BadRequest, "The user don't have permission to update the project's tasks.")
			}
		}
	}

	err := m.TaskService.WithTrx(trx).Delete(&taskModel.Field{
		DeletedTaskUUIDs: input.Tasks,
	})
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// sync delete task_resource
	err = m.syncDeleteTaskResources(trx, nil, input.Tasks, true)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// sync update project's start and end dates
	err = m.syncUpdateProjectStartEndDate(trx, input.ProjectUUID, input.Tasks, nil, nil)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, "Delete ok!")
}

func (m *manager) Update(trx *gorm.DB,
	input *taskModel.Update) (int, any) {
	defer trx.Rollback()

	// get all resources of the project
	proResBase, err := m.ProjectResourceService.GetByListNoPagination(&projectResourceModel.Field{
		ProjectUUID: input.ProjectUUID,
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	var proRes []*projectResourceModel.Single
	proResByte, err := sonic.Marshal(proResBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(proResByte, &proRes)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// create a map of resourceUUID
	proResMap := make(map[string]*projectResourceModel.Single)
	if len(proRes) > 0 {
		for _, res := range proRes {
			proResMap[res.ResourceUUID] = res
		}
	}

	taskBase, err := m.TaskService.GetBySingle(&taskModel.Field{
		TaskUUID: input.TaskUUID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// get the project for the task
	projectBase, err := m.ProjectService.GetBySingle(&projectModel.Field{
		ProjectUUID: *taskBase.ProjectUUID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// check the update_by has the permission to update the project's tasks
	if *input.Role != "admin" {
		if proResMap[*input.ResUUID] != nil {
			if !proResMap[*input.ResUUID].IsEditable {
				log.Info("The user don't have permission to update the project's tasks.")
				return code.BadRequest, code.GetCodeMessage(code.BadRequest, "The user don't have permission to update the project's tasks.")
			}
		}

		// check if the update_by is the task's creator or assigned resource
		if taskBase.CreatedBy != input.UpdatedBy {
			assigned := false
			for _, res := range taskBase.TaskResources {
				if *res.ResourceUUID == *input.ResUUID {
					assigned = true
					break
				}
			}
			if !assigned {
				log.Info("The user don't have permission to update the project's tasks.")
				return code.BadRequest, code.GetCodeMessage(code.BadRequest, "The user don't have permission to update the project's tasks.")
			}
		}
	}

	// determine the project status
	if *projectBase.Status != "建檔中" {
		if *projectBase.Status == "執行中" || *projectBase.Status == "暫停中" {
			if (!input.BaselineStartDate.IsZero() && input.BaselineStartDate != taskBase.BaselineStartDate) || (!input.BaselineEndDate.IsZero() && input.BaselineEndDate != taskBase.BaselineEndDate) {
				log.Info("The baseline cannot be modified while project is in progress.")
				return code.BadRequest, code.GetCodeMessage(code.BadRequest, "The baseline cannot be modified while project is in progress.")
			}
		} else {
			// transform taskBase to update struct
			original := &taskModel.Update{}
			taskByte, err := sonic.Marshal(taskBase)
			if err != nil {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}

			err = sonic.Unmarshal(taskByte, &original)
			if err != nil {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}

			if input != original {
				log.Info("The project is completed and cannot be modified.")
				return code.BadRequest, code.GetCodeMessage(code.BadRequest, "The project is completed and cannot be modified.")
			}
		}
	}

	// get the quantity of subtasks
	subQuantity, _ := m.TaskService.GetByQuantity(&taskModel.Field{
		OutlineNumber: taskBase.OutlineNumber,
	})

	if len(input.Segments) > 0 {
		if subQuantity > 0 {
			log.Info("The parent task cannot be segmented.")
			return code.BadRequest, code.GetCodeMessage(code.BadRequest, "The parent task cannot be segmented.")
		}

		// transform segments from struct array to string
		segJson, _ := sonic.Marshal(input.Segments)
		input.Segment = util.PointerString(string(segJson))
	}

	// transform indicators from struct array to string
	if len(input.Indicators) > 0 {
		indJson, _ := sonic.Marshal(input.Indicators)
		input.Indicator = util.PointerString(string(indJson))
	}

	// sync delete task_resource
	err = m.syncDeleteTaskResources(trx, util.PointerString(input.TaskUUID), nil, false)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// sync create task_resource
	if len(input.Resources) > 0 {
		taskResMapList := []map[string][]*resourceModel.TaskSingle{
			{*taskBase.TaskUUID: input.Resources},
		}
		err = m.syncCreateTaskResources(trx, taskResMapList, *input.UpdatedBy, *input.ProjectUUID, proResMap)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// sync update project's start and end dates
	err = m.syncUpdateProjectStartEndDate(trx, taskBase.ProjectUUID, nil, input.BaselineStartDate, input.BaselineEndDate)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.TaskService.WithTrx(trx).Update(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, taskBase.TaskUUID)
}

func (m *manager) UpdateAll(trx *gorm.DB, input []*taskModel.Update) (int, any) {
	defer trx.Rollback()

	log.Info("UpdateAll Start !!")
	var (
		updateList                       []*taskModel.Update
		taskResMapList                   []map[string][]*resourceModel.TaskSingle
		TaskUUIDs                        []*string
		minBaselineStart, maxBaselineEnd *time.Time
	)

	// get all resources of the project
	proResBase, err := m.ProjectResourceService.GetByListNoPagination(&projectResourceModel.Field{
		ProjectUUID: input[0].ProjectUUID,
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	var proRes []*projectResourceModel.Single
	proResByte, err := sonic.Marshal(proResBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(proResByte, &proRes)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// create a map of resourceUUID
	proResMap := make(map[string]*projectResourceModel.Single)
	if len(proRes) > 0 {
		for _, res := range proRes {
			proResMap[res.ResourceUUID] = res
		}
	}

	// check the update_by has the permission to update the project's tasks
	if *input[0].Role != "admin" {
		if proResMap[*input[0].ResUUID] != nil {
			if !proResMap[*input[0].ResUUID].IsEditable {
				log.Error("The user don't have permission to update the project's tasks.")
				return code.BadRequest, code.GetCodeMessage(code.BadRequest, "The user don't have permission to update the project's tasks.")
			}
		} else {
			log.Error("The user don't have permission to update the project's tasks.")
			return code.BadRequest, code.GetCodeMessage(code.BadRequest, "The user don't have permission to update the project's tasks.")
		}
	}

	// get tasks for the project
	taskBase, err := m.TaskService.GetByListNoPagination(&taskModel.Field{
		ProjectUUID: input[0].ProjectUUID,
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	var tasks []*taskModel.Single
	taskByte, err := sonic.Marshal(taskBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(taskByte, &tasks)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// create a map of taskUUID
	taskMap := make(map[string]*taskModel.Single)
	if len(tasks) > 0 {
		for _, task := range tasks {
			taskMap[task.TaskUUID] = task
		}
	}

	// update the main task
	for i, inputBody := range input {
		// check if the update_by is the task's creator or assigned resource
		if *input[0].Role != "admin" && input[0].ResUUID != nil && taskMap[inputBody.TaskUUID] != nil {
			if taskMap[inputBody.TaskUUID].CreatedBy != *inputBody.UpdatedBy {
				assigned := false
				for _, res := range taskMap[inputBody.TaskUUID].Resources {
					if res.ResourceUUID == *input[0].ResUUID {
						assigned = true
						break
					}
				}
				if !assigned {
					continue
				}
			}
		}

		if inputBody.BaselineStartDate != nil && inputBody.BaselineEndDate != nil {
			// get the minimum baseline_start_date
			if minBaselineStart == nil || inputBody.BaselineStartDate.Before(*minBaselineStart) {
				minBaselineStart = inputBody.BaselineStartDate
			}

			// get the maximum baseline_end_date
			if maxBaselineEnd == nil || inputBody.BaselineEndDate.After(*maxBaselineEnd) {
				maxBaselineEnd = inputBody.BaselineEndDate
			}
		}

		// the main task with subtasks cannot be segmented
		if len(inputBody.Segments) > 0 {
			if len(inputBody.Subtask) > 0 {
				log.Info("The parent task cannot be segmented.")
				return code.BadRequest, code.GetCodeMessage(code.BadRequest, "The parent task cannot be segmented.")
			}

			// transform segments from struct array to string
			segJson, _ := sonic.Marshal(inputBody.Segments)
			inputBody.Segment = util.PointerString(string(segJson))
		}

		// transform indicators from struct array to string
		if len(inputBody.Indicators) > 0 {
			indJson, _ := sonic.Marshal(inputBody.Indicators)
			inputBody.Indicator = util.PointerString(string(indJson))
		}

		inputBody.IsSubTask = util.PointerBool(false)
		// calculate the new outline_number
		inputBody.OutlineNumber = util.PointerString(strconv.Itoa(1 + i))

		// sync update task_resource
		TaskUUIDs = append(TaskUUIDs, util.PointerString(inputBody.TaskUUID))
		if len(inputBody.Resources) > 0 {
			taskResources := make(map[string][]*resourceModel.TaskSingle)
			taskResources[inputBody.TaskUUID] = inputBody.Resources
			taskResMapList = append(taskResMapList, taskResources)
		}

		// update the main task
		updateList = append(updateList, inputBody)

		// update subtasks
		if len(inputBody.Subtask) > 0 {
			subTaskUUIDs, subUpdateList, subTaskResMapList, subMinBaselineStart, subMaxBaselineEnd, err := m.updateSubtasks(trx, inputBody, inputBody.Subtask, minBaselineStart, maxBaselineEnd, input[0].Role, input[0].ResUUID, taskMap)
			if err != nil {
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
			taskResMapList = append(taskResMapList, subTaskResMapList...)
			updateList = append(updateList, subUpdateList...)
			TaskUUIDs = append(TaskUUIDs, subTaskUUIDs...)

			if inputBody.BaselineStartDate != nil && inputBody.BaselineEndDate != nil {
				// compare the minimum baseline_start_date and maximum baseline_end_date of the subtasks
				if minBaselineStart == nil || inputBody.BaselineStartDate.Before(*subMinBaselineStart) {
					minBaselineStart = inputBody.BaselineStartDate
				}
				if maxBaselineEnd == nil || inputBody.BaselineEndDate.After(*subMaxBaselineEnd) {
					maxBaselineEnd = inputBody.BaselineEndDate
				}
			}
		}
	}

	var wg sync.WaitGroup
	// make an error channel
	goroutineErr := make(chan error)
	// create goroutine
	for _, task := range updateList {
		wg.Add(1)
		go func(task *taskModel.Update) {
			defer wg.Done()
			// update task
			err := m.TaskService.WithTrx(trx).Update(task)
			if err != nil {
				log.Error(err)
				goroutineErr <- err
			}
		}(task)
	}

	// wait until all goroutines are finished
	wg.Wait()
	close(goroutineErr)

	// check if goroutine has error
	err = <-goroutineErr
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// sync delete task_resource
	err = m.syncDeleteTaskResources(trx, nil, TaskUUIDs, true)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// sync create task_resource
	if len(taskResMapList) > 0 {
		err = m.syncCreateTaskResources(trx, taskResMapList, *input[0].UpdatedBy, *input[0].ProjectUUID, proResMap)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// sync update project's start and end dates
	err = m.syncUpdateProjectStartEndDate(trx, input[0].ProjectUUID, nil, minBaselineStart, maxBaselineEnd)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, "Successful update!")
}

func (m *manager) Import(trx *gorm.DB, input *taskModel.Import) (int, any) {
	defer trx.Rollback()

	// get resources
	resBase, err := m.ResourceService.GetByListNoPagination(&resourceModel.Field{})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// create a map of resource_name and project_resource_id
	nameToResIDMap := make(map[string]string)
	for _, res := range resBase {
		nameToResIDMap[*res.ResourceName] = *res.ResourceUUID
	}

	// set CSV parse options
	input.CSVFile.LazyQuotes = true       // loosely process quotes
	input.CSVFile.TrimLeadingSpace = true // automatically remove spaces before each field
	input.CSVFile.FieldsPerRecord = -1    // do not force each record to have the same number of fields

	// read and parse CSV file
	records, err := input.CSVFile.ReadAll()
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	taskIdx := [18]int{}
	taskRecordIdx := make(map[string]int)
	var createAllTask []*taskModel.Create
	for i, record := range records {
		if i == 0 {
			// identify the CSV header row and record the index of each field
			for index, value := range record {
				log.Debug("index: ", index, " value: ", value)
				// 0: gantt project
				if input.FileType == 1 {
					switch value {
					// set the index of each field according to the column name
					case "ID", "編號":
						taskIdx[0] = index
					case "Name", "名稱":
						taskIdx[1] = index
					case "Begin date", "實際起始日":
						taskIdx[2] = index
					case "End date", "實際完成日":
						taskIdx[3] = index
					case "Duration", "期間":
						taskIdx[4] = index
					case "Completion", "完成":
						taskIdx[5] = index
					case "Cost":
						taskIdx[6] = index
					//case "Coordinator", "協調者":
					//	taskIdx[7] = index
					case "Predecessors", "父階":
						taskIdx[8] = index
					case "Outline number", "大綱編號":
						taskIdx[9] = index
					case "Resources", "資源":
						taskIdx[10] = index
					case "Assignments":
						taskIdx[11] = index
					case "Task color", "任務顏色":
						taskIdx[12] = index
					case "Web Link", "超連結":
						taskIdx[13] = index
					case "Notes", "備註":
						taskIdx[14] = index
					case "Baseline Begin date", "起始日期":
						taskIdx[15] = index
					case "Baseline End date", "結束日期":
						taskIdx[16] = index
					}
					// 1: saas pmi
				} else if input.FileType == 2 {
					switch value {
					// set the index of each field according to the column name
					case "ID", "編號":
						taskIdx[0] = index
					case "Name", "名稱":
						taskIdx[1] = index
					case "Begin date", "開始日期":
						taskIdx[2] = index
					case "End date", "結束日期":
						taskIdx[3] = index
					case "Duration", "工作天":
						taskIdx[4] = index
					case "Completion", "進度(%)":
						taskIdx[5] = index
					case "Cost", "工時表":
						taskIdx[6] = index
					case "Predecessors", "相依性":
						taskIdx[8] = index
					case "Outline number", "大綱編號":
						taskIdx[9] = index
					case "Resources", "負責人":
						taskIdx[10] = index
					case "Notes", "備註":
						taskIdx[14] = index
					case "Baseline Begin date", "基準開始日":
						taskIdx[15] = index
					case "Baseline End date", "基準結束日":
						taskIdx[16] = index
					case "Baseline Duration", "基準工作天":
						taskIdx[17] = index
					}
				}
			}
			continue
		}

		// skip the record if there is no outline_number
		if record[taskIdx[9]] == "" {
			continue
		}

		// build the top outline_number for validation
		outlineNumberSplit := strings.Split(record[taskIdx[9]], ".")
		var topOutlineNumber string
		for i := 0; i < len(outlineNumberSplit)-1; i++ {
			topOutlineNumber += outlineNumberSplit[i]
			if i != len(outlineNumberSplit)-2 {
				topOutlineNumber += "."
			}
		}

		// verify the hierarchy relationship in the file
		if _, ok := taskRecordIdx[topOutlineNumber]; topOutlineNumber != "" && !ok {
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, "import errors")
		}

		// combine the parsed data into the 'Create' structure
		createTask := &taskModel.Create{}
		createTask.TaskID = record[taskIdx[0]]
		createTask.TaskName = record[taskIdx[1]]
		if record[taskIdx[2]] != "" {
			// transform the date format from "/" to "-"
			var startDateString string
			for _, v := range strings.Split(strings.ReplaceAll(record[taskIdx[2]], "/", "-"), "-") {
				if len(v) == 1 {
					startDateString += "0" + v + "-"
				} else {
					startDateString += v + "-"
				}
			}
			startDateString = startDateString[0 : len(startDateString)-1]
			startDate, err := time.Parse("2006-01-02", startDateString)
			if err != nil {
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
			createTask.StartDate = util.PointerTime(startDate)
		}

		if record[taskIdx[3]] != "" {
			// transform the date format from "/" to "-"
			var endDateString string
			for _, v := range strings.Split(strings.ReplaceAll(record[taskIdx[3]], "/", "-"), "-") {
				if len(v) == 1 {
					endDateString += "0" + v + "-"
				} else {
					endDateString += v + "-"
				}
			}
			endDateString = endDateString[0 : len(endDateString)-1]
			endDateString += " 12:00:00"
			endDate, err := time.Parse("2006-01-02 15:04:05", endDateString)
			if err != nil {
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
			createTask.EndDate = util.PointerTime(endDate)
		}

		//if record[taskIdx[4]] != "" {
		//	duration, err := strconv.ParseFloat(record[taskIdx[4]], 64)
		//	if err != nil {
		//		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		//	}
		//	createTask.Duration = duration
		//}

		if record[taskIdx[5]] != "" {
			progress, _ := strconv.Atoi(record[taskIdx[5]])
			createTask.Progress = int64(progress)
		}

		if record[taskIdx[6]] != "" {
			cost, err := strconv.ParseFloat(record[taskIdx[6]], 64)
			if err != nil {
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
			createTask.Cost = int64(cost)
		}

		createTask.Predecessor = record[taskIdx[8]]
		if record[taskIdx[10]] != "" {
			var resourceSplit []string
			separators := []string{";", ","}
			for _, sep := range separators {
				if strings.Contains(record[taskIdx[10]], sep) {
					resourceSplit = strings.Split(record[taskIdx[10]], sep)
					break
				}
			}

			if len(resourceSplit) == 0 {
				resourceSplit = append(resourceSplit, record[taskIdx[10]])
			}

			for _, res := range resourceSplit {
				var percentage float64
				var name string
				if strings.Contains(res, "[") {
					res = strings.Trim(res, "[]")
					parts := strings.Split(res, "[")
					name = parts[0]
					percentage, _ = strconv.ParseFloat(strings.TrimRight(parts[1], "%"), 64)
				} else {
					name = res
					percentage = 100
				}

				if nameToResIDMap[name] != "" {
					createTask.Resources = append(createTask.Resources, &resourceModel.TaskSingle{
						ResourceUUID: nameToResIDMap[name],
						Unit:         percentage,
					})
				}
				continue
			}
		}

		createTask.Assignments = record[taskIdx[11]]
		createTask.TaskColor = record[taskIdx[12]]
		createTask.WebLink = record[taskIdx[13]]
		createTask.Notes = record[taskIdx[14]]

		if record[taskIdx[15]] != "" {
			// transform the date format from "/" to "-"
			var startDateString string
			for _, v := range strings.Split(strings.ReplaceAll(record[taskIdx[15]], "/", "-"), "-") {
				if len(v) == 1 {
					startDateString += "0" + v + "-"
				} else {
					startDateString += v + "-"
				}
			}
			startDateString = startDateString[0 : len(startDateString)-1]
			startDate, err := time.Parse("2006-01-02", startDateString)
			if err != nil {
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
			createTask.BaselineStartDate = util.PointerTime(startDate)
		}

		if record[taskIdx[16]] != "" {
			// transform the date format from "/" to "-"
			var endDateString string
			for _, v := range strings.Split(strings.ReplaceAll(record[taskIdx[16]], "/", "-"), "-") {
				if len(v) == 1 {
					endDateString += "0" + v + "-"
				} else {
					endDateString += v + "-"
				}
			}
			endDateString = endDateString[0 : len(endDateString)-1]
			endDateString += " 12:00:00"
			endDate, err := time.Parse("2006-01-02 15:04:05", endDateString)
			if err != nil {
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
			createTask.BaselineEndDate = util.PointerTime(endDate)
		}

		if record[taskIdx[17]] != "" {
			duration := strings.Replace(record[taskIdx[17]], "天", "", -1)
			durationNum, err := strconv.ParseFloat(duration, 64)
			if err != nil {
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
			createTask.BaselineDuration = durationNum
		}

		// check if there is no data with the same outline_number
		if _, ok := taskRecordIdx[record[taskIdx[9]]]; !ok {
			// record the index of the current task in createAllTask
			taskRecordIdx[record[taskIdx[9]]] = len(createAllTask)
		}

		createTask.CreatedBy = input.CreatedBy
		createAllTask = assembleToCreateAll(createAllTask, taskRecordIdx, createTask, record[taskIdx[9]], "", input.ProjectUUID, input.ResUUID, input.Role)
	}

	httpCode, message := m.CreateAll(trx, createAllTask)
	if httpCode != code.Successful {
		return httpCode, message
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, "Successful import!")
}
