package task

import (
	"encoding/json"
	"errors"
	"fmt"
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
			// 取得最小的baseline_start_date
			if minBaselineStart == nil || subBody.BaselineStartDate.Before(*minBaselineStart) {
				minBaselineStart = subBody.BaselineStartDate
			}

			// 取得最大的baseline_end_date
			if maxBaselineEnd == nil || subBody.BaselineEndDate.After(*maxBaselineEnd) {
				maxBaselineEnd = subBody.BaselineEndDate
			}
		}

		newOutlineNumber := fmt.Sprintf("%s.%d", parentTask.OutlineNumber, j+1)
		subBody.OutlineNumber = newOutlineNumber
		subBody.IsSubTask = true
		subBody.CreatedBy = parentTask.CreatedBy
		subBody.ProjectUUID = parentTask.ProjectUUID
		// 將segments轉換為JSON
		if len(subBody.Segments) > 0 {
			// 確認是否為最子層任務
			if len(subBody.Subtask) > 0 {
				return nil, nil, nil, errors.New("the parent task cannot be segmented")
			}
			segJson, err := json.Marshal(subBody.Segments)
			if err != nil {
				log.Error(err)
				return nil, nil, nil, err
			}
			subBody.Segment = string(segJson)
		}

		// 將indicators轉換為JSON
		if len(subBody.Indicators) > 0 {
			indJson, err := json.Marshal(subBody.Indicators)
			if err != nil {
				log.Error(err)
				return nil, nil, nil, err
			}
			subBody.Indicator = string(indJson)
		}
		createList = append(createList, subBody)

		// 處理可能的子任務
		if len(subBody.Subtask) > 0 {
			subSubtasks, subMinBaselineStart, subMaxBaselineEnd, err := m.createSubtasks(trx, subBody, subBody.Subtask, projectStart, projectEnd, minBaselineStart, maxBaselineEnd)
			if err != nil {
				return nil, nil, nil, err
			}
			createList = append(createList, subSubtasks...)

			// 比對子任務的最小baseline_start_date及最大baseline_end_date
			if minBaselineStart == nil || subBody.BaselineStartDate.Before(*subMinBaselineStart) {
				minBaselineStart = subBody.BaselineStartDate
			}
			if maxBaselineEnd == nil || subBody.BaselineEndDate.After(*subMaxBaselineEnd) {
				maxBaselineEnd = subBody.BaselineEndDate
			}
		}
	}

	return createList, minBaselineStart, maxBaselineEnd, nil
}

// updateSubtasks is a helper function to update subtasks recursively for a parent task and returns the UUIDs of updated subtasks.
func (m *manager) updateSubtasks(trx *gorm.DB, parentTask *taskModel.Update, subtasks []*taskModel.Update, minBaselineStart, maxBaselineEnd *time.Time) ([]*string, []*taskModel.Update, []map[string][]*resourceModel.TaskSingle, *time.Time, *time.Time, error) {
	var (
		taskResMapList []map[string][]*resourceModel.TaskSingle
		updateList     []*taskModel.Update
		TaskUUIDs      []*string
	)

	for j, subBody := range subtasks {
		if subBody.BaselineStartDate != nil && subBody.BaselineEndDate != nil {
			// 取得最小的baseline_start_date
			if minBaselineStart == nil || subBody.BaselineStartDate.Before(*minBaselineStart) {
				minBaselineStart = subBody.BaselineStartDate
			}

			// 取得最大的baseline_end_date
			if maxBaselineEnd == nil || subBody.BaselineEndDate.After(*maxBaselineEnd) {
				maxBaselineEnd = subBody.BaselineEndDate
			}
		}

		subBody.UpdatedBy = parentTask.UpdatedBy
		subBody.ProjectUUID = parentTask.ProjectUUID
		subBody.IsSubTask = util.PointerBool(true)
		// 更新任務階層
		newOutlineNumber := fmt.Sprintf("%s.%d", *parentTask.OutlineNumber, j+1)
		subBody.OutlineNumber = util.PointerString(newOutlineNumber)

		// 將segments轉換為JSON
		if len(subBody.Segments) > 0 {
			// 確認是否為最子層任務
			if len(subBody.Subtask) > 0 {
				return nil, nil, nil, nil, nil, errors.New("the parent task cannot be segmented")
			}
			segJson, err := json.Marshal(subBody.Segments)
			if err != nil {
				log.Error(err)
				return nil, nil, nil, nil, nil, err
			}
			subBody.Segment = util.PointerString(string(segJson))
		} else {
			subBody.Segment = nil
		}

		// 將indicators轉換為JSON
		if len(subBody.Indicators) > 0 {
			indJson, err := json.Marshal(subBody.Indicators)
			if err != nil {
				log.Error(err)
				return nil, nil, nil, nil, nil, err
			}
			subBody.Indicator = util.PointerString(string(indJson))
		} else {
			subBody.Indicator = nil
		}

		// 更新task_resource
		if len(subBody.Resources) > 0 {
			TaskUUIDs = append(TaskUUIDs, util.PointerString(subBody.TaskUUID))
			taskResources := make(map[string][]*resourceModel.TaskSingle)
			taskResources[subBody.TaskUUID] = subBody.Resources
			taskResMapList = append(taskResMapList, taskResources)
		}

		// 子任務更新
		updateList = append(updateList, subBody)

		// 處理可能的子任務
		if len(subBody.Subtask) > 0 {
			subTaskUUIDs, subUpdateList, subTaskResMapList, subMinBaselineStart, subMaxBaselineEnd, err := m.updateSubtasks(trx, subBody, subBody.Subtask, minBaselineStart, maxBaselineEnd)
			if err != nil {
				return nil, nil, nil, nil, nil, err
			}
			taskResMapList = append(taskResMapList, subTaskResMapList...)
			updateList = append(updateList, subUpdateList...)
			TaskUUIDs = append(TaskUUIDs, subTaskUUIDs...)

			if subBody.BaselineStartDate != nil && subBody.BaselineEndDate != nil {
				// 比對子任務的最小baseline_start_date及最大baseline_end_date
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
func (m *manager) syncCreateTaskResources(trx *gorm.DB, taskResources []map[string][]*resourceModel.TaskSingle, createdBy string) error {
	var resList []*taskResourceModel.Create

	for _, taskRes := range taskResources {
		for taskUUID, resources := range taskRes {
			for _, res := range resources {
				resList = append(resList, &taskResourceModel.Create{
					TaskUUID:     taskUUID,
					ResourceUUID: res.ResourceUUID,
					Unit:         res.Unit,
					CreatedBy:    createdBy,
				})
			}
		}
	}

	_, err := m.TaskResourceService.WithTrx(trx).CreateAll(resList)
	if err != nil {
		return err
	}
	return nil
}

// syncDeleteTaskResources is a helper function to synchronize the deletion of task_resource associations for a tasks of a project.
func (m *manager) syncDeleteTaskResources(trx *gorm.DB, TaskUUID *string, TaskUUIDs []*string, isBatch bool) error {
	if isBatch {
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
	// 取得任務最早開始日期及最晚結束日期
	taskBase, err := m.TaskService.GetByMinStartMaxEnd(&taskModel.Field{
		ProjectUUID:      projectID,
		DeletedTaskUUIDs: TaskUUIDs,
	})
	if err != nil {
		return err
	}

	if len(taskBase) > 0 {
		// 找到taskBase中最小的baseline_start_date及最大的baseline_end_date
		for _, task := range taskBase {
			if minBaselineStart == nil || task.BaselineStartDate.Before(*minBaselineStart) {
				minBaselineStart = task.BaselineStartDate
			}
			if maxBaselineEnd == nil || task.BaselineEndDate.After(*maxBaselineEnd) {
				maxBaselineEnd = task.BaselineEndDate
			}
		}
	}

	// 比對最新資料最小的baseline_start_date及最大的baseline_end_date
	if start != nil && end != nil {
		if minBaselineStart == nil || start.Before(*minBaselineStart) {
			minBaselineStart = start
		}
		if maxBaselineEnd == nil || end.After(*maxBaselineEnd) {
			maxBaselineEnd = end
		}
	}

	// 更新專案開始日期及結束日期
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
func assembleToCreateAll(tasks []*taskModel.Create, outlineNumber string, selectOutlineNumber string, taskRecordIdx map[string]int, task *taskModel.Create, projectID string) []*taskModel.Create {
	task.ProjectUUID = projectID
	outlineNumberSplit := strings.Split(outlineNumber, ".")

	//主任務處理
	if len(outlineNumberSplit) == 1 {
		//紀錄當前主任務的outline_number
		if outlineNumber != "" {
			selectOutlineNumber += "." + outlineNumber
		}
		// 將當前主任務的outline_number映射到allTask的index位置
		taskRecordIdx[selectOutlineNumber] = len(tasks)
		tasks = append(tasks, task)
		return tasks
	} else {
		//子任務處理
		var newOutlineNumber string
		for i := 1; i < len(outlineNumberSplit); i++ {
			newOutlineNumber += outlineNumberSplit[i]
			//如果還不是最後的子任務，將outline_number加入分隔符
			if i != len(outlineNumberSplit)-1 {
				newOutlineNumber += "."
			}
		}

		//將當前子任務的outline_number添加到其主任務的outline_number後
		if selectOutlineNumber == "" {
			selectOutlineNumber = outlineNumberSplit[0]
		} else {
			selectOutlineNumber += "." + outlineNumberSplit[0]
		}

		//處理可能的子任務
		subTasks := assembleToCreateAll(tasks[taskRecordIdx[selectOutlineNumber]].Subtask, newOutlineNumber, selectOutlineNumber, taskRecordIdx, task, projectID)
		tasks[taskRecordIdx[selectOutlineNumber]].Subtask = subTasks
	}

	return tasks
}

func (m *manager) Create(trx *gorm.DB, input *taskModel.Create) (int, any) {
	defer trx.Rollback()

	// 判斷是否為子任務
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

		// 取得同階層最後的outline_number
		input.OutlineNumber, err = m.getNextOutlineNumber(true, parentBase.OutlineNumber, input.ProjectUUID)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
		input.IsSubTask = true

		// 確認是否為最子層任務
		if parentBase.Segment != nil {
			if *parentBase.Segment != "" {
				log.Info("Please remove the task segmentation first.")
				return code.BadRequest, code.GetCodeMessage(code.BadRequest, "Please remove the task segmentation first.")
			}
		}

		// 將segments轉換為JSON
		if len(input.Segments) > 0 {
			segJson, _ := json.Marshal(input.Segments)
			input.Segment = string(segJson)
		}

		// 將indicators轉換為JSON
		if len(input.Indicators) > 0 {
			indJson, _ := json.Marshal(input.Indicators)
			input.Indicator = string(indJson)
		}

	} else {
		// 有子任務的主任務不能分段
		if len(input.Segments) > 0 {
			if len(input.Subtask) > 0 {
				log.Info("The parent task cannot be segmented.")
				return code.BadRequest, code.GetCodeMessage(code.BadRequest, "The parent task cannot be segmented.")
			}
			// 將segments轉換為JSON
			segJson, _ := json.Marshal(input.Segments)
			input.Segment = string(segJson)
		}

		// 將indicators轉換為JSON
		if len(input.Indicators) > 0 {
			indJson, _ := json.Marshal(input.Indicators)
			input.Indicator = string(indJson)
		}

		// 取得新outline number
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

	// 同步新增task_resource關聯
	if len(input.Resources) > 0 {
		taskResMapList := []map[string][]*resourceModel.TaskSingle{
			{*taskBase.TaskUUID: input.Resources},
		}
		err = m.syncCreateTaskResources(trx, taskResMapList, input.CreatedBy)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// 同步修改專案開始日期及結束日期
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
	// 取得最新outline_number
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

	// 主任務新增
	for i, inputBody := range input {
		if inputBody.BaselineStartDate != nil && inputBody.BaselineEndDate != nil {
			// 取得最小的baseline_start_date
			if minBaselineStart == nil || inputBody.BaselineStartDate.Before(*minBaselineStart) {
				minBaselineStart = inputBody.BaselineStartDate
			}

			// 取得最大的baseline_end_date
			if maxBaselineEnd == nil || inputBody.BaselineEndDate.After(*maxBaselineEnd) {
				maxBaselineEnd = inputBody.BaselineEndDate
			}
		}

		// 有子任務的主任務不能分段
		if len(inputBody.Segments) > 0 {
			if len(inputBody.Subtask) > 0 {
				log.Info("The parent task cannot be segmented.")
				return code.BadRequest, code.GetCodeMessage(code.BadRequest, "The parent task cannot be segmented.")
			}
			// 將segments轉換為JSON
			segJson, _ := json.Marshal(inputBody.Segments)
			inputBody.Segment = string(segJson)
		}

		// 將indicators轉換為JSON
		if len(inputBody.Indicators) > 0 {
			indJson, _ := json.Marshal(inputBody.Indicators)
			inputBody.Indicator = string(indJson)
		}

		// 取得新outline number
		inputBody.OutlineNumber = strconv.Itoa(newOutlineNumber + i)
		createList = append(createList, inputBody)

		// 子任務新增
		if len(inputBody.Subtask) > 0 {
			subSubtasks, subMinBaselineStart, subMaxBaselineEnd, err := m.createSubtasks(trx, inputBody, inputBody.Subtask, projectBase.StartDate, projectBase.EndDate, minBaselineStart, maxBaselineEnd)
			if err != nil {
				if err.Error() == "the parent task cannot be segmented" {
					return code.BadRequest, code.GetCodeMessage(code.BadRequest, "The parent task cannot be segmented.")
				}

				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
			createList = append(createList, subSubtasks...)

			// 比對子任務的最小baseline_start_date及最大baseline_end_date
			if minBaselineStart == nil || inputBody.BaselineStartDate.Before(*subMinBaselineStart) {
				minBaselineStart = inputBody.BaselineStartDate
			}
			if maxBaselineEnd == nil || inputBody.BaselineEndDate.After(*subMaxBaselineEnd) {
				maxBaselineEnd = inputBody.BaselineEndDate
			}
		}
	}

	tasksBase, err := m.TaskService.WithTrx(trx).CreateAll(createList)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// 同步新增task_resource關聯
	for i, taskBase := range tasksBase {
		if len(createList[i].Resources) > 0 {
			taskResources := make(map[string][]*resourceModel.TaskSingle)
			taskResources[*taskBase.TaskUUID] = createList[i].Resources
			taskResMapList = append(taskResMapList, taskResources)
		}
	}
	if len(taskResMapList) > 0 {
		err = m.syncCreateTaskResources(trx, taskResMapList, createList[0].CreatedBy)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// 同步修改專案開始日期及結束日期
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

	// 取得project資訊
	projectBase, err := m.ProjectService.GetByListNoPagination(&projectModel.Field{
		ProjectIDs: input.Projects,
	})
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	var projects []*projectModel.Single
	projectByte, err := json.Marshal(projectBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = json.Unmarshal(projectByte, &projects)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// 建立projectUUID的映射表
	projectMap := make(map[string]*projectModel.Single)
	for _, project := range projects {
		projectMap[project.ProjectUUID] = project
	}

	// 取得projectTasks資訊
	projectTaskMap := make(map[*string][]*taskModel.Single)
	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)
	// 建立error channel
	goroutineErr := make(chan error)
	// 建立goroutine
	for _, projectsUUID := range input.Projects {
		wg.Add(1)
		go func(projectsUUID *string) {
			defer wg.Done()
			mu.Lock()
			defer mu.Unlock()

			// 取得task資訊
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
			taskByte, err := json.Marshal(taskBase)
			if err != nil {
				log.Error(err)
				goroutineErr <- err
			}

			err = json.Unmarshal(taskByte, &projectTasks)
			if err != nil {
				log.Error(err)
				goroutineErr <- err
			}

			// 建立任務UUID的映射表
			taskMap := make(map[string]*taskModel.Single)
			for _, task := range projectTasks {
				taskMap[task.TaskUUID] = task
			}

			for i, task := range projectTasks {
				task.CreatedBy = *taskBase[i].CreatedByUsers.Name
				task.UpdatedBy = *taskBase[i].UpdatedByUsers.Name

				// 將segment轉為陣列
				var segments []taskModel.Segments
				err = util.DecodeJSONToSlice(*taskBase[i].Segment, &segments)
				if err != nil {
					log.Error(err)
					goroutineErr <- err
				}
				task.Segments = segments

				// 將indicator轉為陣列
				var indicators []taskModel.Indicators
				err = util.DecodeJSONToSlice(*taskBase[i].Indicator, &indicators)
				if err != nil {
					log.Error(err)
					goroutineErr <- err
				}
				task.Indicators = indicators

				task.CoordinatorName = *taskBase[i].Coordinators.ResourceName
				for j, res := range taskBase[i].TaskResources {
					task.Resources[j].ResourceUUID = *res.Resources.ResourceUUID
					task.Resources[j].ResourceID = *res.Resources.Resources.ResourceID
					task.Resources[j].ResourceName = *res.Resources.Resources.ResourceName
					task.Resources[j].Email = *res.Resources.Resources.Email
					task.Resources[j].Phone = *res.Resources.Resources.Phone
					task.Resources[j].StandardCost = *res.Resources.Resources.StandardCost
					task.Resources[j].TotalCost = *res.Resources.Resources.TotalCost
					task.Resources[j].TotalLoad = *res.Resources.Resources.TotalLoad
					task.Resources[j].ResourceGroup = *res.Resources.Resources.ResourceGroup
					task.Resources[j].IsExpand = *res.Resources.Resources.IsExpand
					task.Resources[j].Role = *res.Resources.Role
				}

				if !input.FilterMilestone {
					// 判斷是否為子任務
					if task.IsSubTask {
						// 取得parent_outline_number
						parentOutlineNumber := getParentOutlineNumber(*taskBase[i].OutlineNumber)
						if parentOutlineNumber != "" {
							// 透過parent_outline_number尋找相同父階層的任務
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

	// 等待所有goroutine完成
	wg.Wait()
	close(goroutineErr)

	// 檢查錯誤
	err = <-goroutineErr
	if err != nil {
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	for _, projectsUUID := range input.Projects {
		if len(projectTaskMap[projectsUUID]) > 0 {
			if len(input.Projects) == 1 {
				if !input.FilterMilestone {
					// 過濾掉已經列入各個SubTask的子任務
					var filteredTasks []*taskModel.Single
					for _, task := range projectTaskMap[projectsUUID] {
						if !task.IsSubTask {
							filteredTasks = append(filteredTasks, task)
						}
					}
					output.Tasks = filteredTasks
				}

			} else {
				projectTask := &taskModel.Single{
					TaskUUID:  projectMap[*projectsUUID].ProjectUUID,
					TaskID:    projectMap[*projectsUUID].ProjectID,
					TaskName:  projectMap[*projectsUUID].ProjectName,
					IsSubTask: false,
				}

				projectTask.Subtask = projectTaskMap[projectsUUID]

				if !input.FilterMilestone {
					// 過濾掉已經列入各個SubTask的子任務
					var filteredTasks []*taskModel.Single
					for _, task := range projectTask.Subtask {
						task.Predecessor = ""
						if !task.IsSubTask {
							filteredTasks = append(filteredTasks, task)
						}
					}
					projectTask.Subtask = filteredTasks
				}

				// 將project加入output.Tasks
				output.Tasks = append(output.Tasks, projectTask)
			}

			// 取得event_marks
			var eventMarks []*eventMarkModel.Single
			eventMarkBase, err := m.EventMarkService.GetByListNoPagination(&eventMarkModel.Field{
				ProjectUUID: projectsUUID,
			})
			if err != nil {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
			eventMarkByte, err := json.Marshal(eventMarkBase)
			if err != nil {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}

			err = json.Unmarshal(eventMarkByte, &eventMarks)
			if err != nil {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}

			// 將project加入output.Tasks
			output.EventMarks = append(output.EventMarks, eventMarks...)
		}
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

	taskByte, err := json.Marshal(taskBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = json.Unmarshal(taskByte, &output.Tasks)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	for i, task := range output.Tasks {
		task.CreatedBy = *taskBase[i].CreatedByUsers.Name
		task.UpdatedBy = *taskBase[i].UpdatedByUsers.Name

		// 將segment轉為陣列
		var segments []taskModel.Segments
		err = util.DecodeJSONToSlice(*taskBase[i].Segment, &segments)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
		task.Segments = segments

		// 將indicator轉為陣列
		var indicators []taskModel.Indicators
		err = util.DecodeJSONToSlice(*taskBase[i].Indicator, &indicators)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
		task.Indicators = indicators

		task.CoordinatorName = *taskBase[i].Coordinators.ResourceName
		for j, res := range taskBase[i].TaskResources {
			task.Resources[j].ResourceUUID = *res.Resources.ResourceUUID
			task.Resources[j].ResourceID = *res.Resources.Resources.ResourceID
			task.Resources[j].ResourceName = *res.Resources.Resources.ResourceName
			task.Resources[j].Email = *res.Resources.Resources.Email
			task.Resources[j].Phone = *res.Resources.Resources.Phone
			task.Resources[j].StandardCost = *res.Resources.Resources.StandardCost
			task.Resources[j].TotalCost = *res.Resources.Resources.TotalCost
			task.Resources[j].TotalLoad = *res.Resources.Resources.TotalLoad
			task.Resources[j].ResourceGroup = *res.Resources.Resources.ResourceGroup
			task.Resources[j].IsExpand = *res.Resources.Resources.IsExpand
			task.Resources[j].Role = *res.Resources.Role
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
	taskByte, _ := json.Marshal(taskBase)
	err = json.Unmarshal(taskByte, &output)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output.CreatedBy = *taskBase.CreatedByUsers.Name
	output.UpdatedBy = *taskBase.UpdatedByUsers.Name

	// 將segment轉為陣列
	var segments []taskModel.Segments
	err = util.DecodeJSONToSlice(*taskBase.Segment, &segments)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	output.Segments = segments

	// 將indicator轉為陣列
	var indicators []taskModel.Indicators
	err = util.DecodeJSONToSlice(*taskBase.Indicator, &indicators)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	output.Indicators = indicators

	output.CoordinatorName = *taskBase.Coordinators.ResourceName
	for i, res := range taskBase.TaskResources {
		output.Resources[i].ResourceUUID = *res.Resources.ResourceUUID
		output.Resources[i].ResourceID = *res.Resources.Resources.ResourceID
		output.Resources[i].ResourceName = *res.Resources.Resources.ResourceName
		output.Resources[i].Email = *res.Resources.Resources.Email
		output.Resources[i].Phone = *res.Resources.Resources.Phone
		output.Resources[i].StandardCost = *res.Resources.Resources.StandardCost
		output.Resources[i].TotalCost = *res.Resources.Resources.TotalCost
		output.Resources[i].TotalLoad = *res.Resources.Resources.TotalLoad
		output.Resources[i].ResourceGroup = *res.Resources.Resources.ResourceGroup
		output.Resources[i].IsExpand = *res.Resources.Resources.IsExpand
		output.Resources[i].Role = *res.Resources.Role
	}
	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) Delete(trx *gorm.DB, input *taskModel.DeletedTaskUUIDs) (int, any) {
	defer trx.Rollback()

	err := m.TaskService.WithTrx(trx).Delete(&taskModel.Field{
		DeletedTaskUUIDs: input.Tasks,
	})
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// 同步刪除task_resource關聯
	err = m.syncDeleteTaskResources(trx, nil, input.Tasks, true)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// 同步修改專案開始日期及結束日期
	err = m.syncUpdateProjectStartEndDate(trx, util.PointerString(input.ProjectUUID), input.Tasks, nil, nil)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, "Delete ok!")
}

func (m *manager) Update(trx *gorm.DB, input *taskModel.Update) (int, any) {
	defer trx.Rollback()

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

	// 取得子任務數量
	subQuantity, _ := m.TaskService.GetByQuantity(&taskModel.Field{
		OutlineNumber: taskBase.OutlineNumber,
	})

	if len(input.Segments) > 0 {
		if subQuantity > 0 {
			log.Info("The parent task cannot be segmented.")
			return code.BadRequest, code.GetCodeMessage(code.BadRequest, "The parent task cannot be segmented.")
		}
		// 將segments轉換為JSON
		segJson, _ := json.Marshal(input.Segments)
		input.Segment = util.PointerString(string(segJson))
	}

	if !input.BaselineEndDate.IsZero() || !input.BaselineEndDate.IsZero() {
		// 取得project資訊
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

		if *projectBase.Status == "啟動中" {
			log.Info("The baseline cannot be modified while project is in progress.")
			return code.BadRequest, code.GetCodeMessage(code.BadRequest, "The baseline cannot be modified while project is in progress.")
		}
	}

	// 將indicators轉換為JSON
	if len(input.Indicators) > 0 {
		indJson, _ := json.Marshal(input.Indicators)
		input.Indicator = util.PointerString(string(indJson))
	}

	// 同步刪除task_resource關聯
	err = m.syncDeleteTaskResources(trx, util.PointerString(input.TaskUUID), nil, false)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// 同步新增task_resource關聯
	if len(input.Resources) > 0 {
		taskResMapList := []map[string][]*resourceModel.TaskSingle{
			{*taskBase.TaskUUID: input.Resources},
		}
		err = m.syncCreateTaskResources(trx, taskResMapList, *input.UpdatedBy)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// 同步修改專案開始日期及結束日期
	err = m.syncUpdateProjectStartEndDate(trx, input.ProjectUUID, nil, input.BaselineStartDate, input.BaselineEndDate)
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

	var (
		updateList                       []*taskModel.Update
		taskResMapList                   []map[string][]*resourceModel.TaskSingle
		TaskUUIDs                        []*string
		minBaselineStart, maxBaselineEnd *time.Time
	)

	// 主任務更新
	for i, inputBody := range input {
		if inputBody.BaselineStartDate != nil && inputBody.BaselineEndDate != nil {
			// 取得最小的baseline_start_date
			if minBaselineStart == nil || inputBody.BaselineStartDate.Before(*minBaselineStart) {
				minBaselineStart = inputBody.BaselineStartDate
			}

			// 取得最大的baseline_end_date
			if maxBaselineEnd == nil || inputBody.BaselineEndDate.After(*maxBaselineEnd) {
				maxBaselineEnd = inputBody.BaselineEndDate
			}
		}

		// 有子任務的主任務不能分段
		if len(inputBody.Segments) > 0 {
			if len(inputBody.Subtask) > 0 {
				log.Info("The parent task cannot be segmented.")
				return code.BadRequest, code.GetCodeMessage(code.BadRequest, "The parent task cannot be segmented.")
			}
			// 將segments轉換為JSON
			segJson, _ := json.Marshal(inputBody.Segments)
			inputBody.Segment = util.PointerString(string(segJson))
		}

		// 將indicators轉換為JSON
		if len(inputBody.Indicators) > 0 {
			indJson, _ := json.Marshal(inputBody.Indicators)
			inputBody.Indicator = util.PointerString(string(indJson))
		}

		inputBody.IsSubTask = util.PointerBool(false)
		// 計算薪階層
		inputBody.OutlineNumber = util.PointerString(strconv.Itoa(1 + i))

		// 更新task_resource
		TaskUUIDs = append(TaskUUIDs, util.PointerString(inputBody.TaskUUID))
		if len(inputBody.Resources) > 0 {
			taskResources := make(map[string][]*resourceModel.TaskSingle)
			taskResources[inputBody.TaskUUID] = inputBody.Resources
			taskResMapList = append(taskResMapList, taskResources)
		}

		// 主任務更新
		updateList = append(updateList, inputBody)

		// 子任務更新
		if len(inputBody.Subtask) > 0 {
			subTaskUUIDs, subUpdateList, subTaskResMapList, subMinBaselineStart, subMaxBaselineEnd, err := m.updateSubtasks(trx, inputBody, inputBody.Subtask, minBaselineStart, maxBaselineEnd)
			if err != nil {
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
			taskResMapList = append(taskResMapList, subTaskResMapList...)
			updateList = append(updateList, subUpdateList...)
			TaskUUIDs = append(TaskUUIDs, subTaskUUIDs...)

			if inputBody.BaselineStartDate != nil && inputBody.BaselineEndDate != nil {
				// 比對子任務的最小baseline_start_date及最大baseline_end_date
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
	// 建立error channel
	goroutineErr := make(chan error)
	// 建立goroutine
	for _, task := range updateList {
		wg.Add(1)
		go func(task *taskModel.Update) {
			defer wg.Done()
			// 更新任務
			err := m.TaskService.WithTrx(trx).Update(task)
			if err != nil {
				goroutineErr <- err
			}
		}(task)
	}

	// 等待所有goroutine完成
	wg.Wait()
	close(goroutineErr)

	// 檢查錯誤
	err := <-goroutineErr
	if err != nil {
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// 同步刪除task_resource關聯
	err = m.syncDeleteTaskResources(trx, nil, TaskUUIDs, true)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// 同步新增task_resource關聯
	if len(taskResMapList) > 0 {
		err = m.syncCreateTaskResources(trx, taskResMapList, *input[0].UpdatedBy)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// 同步修改專案開始日期及結束日期
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

	// 取得專案資源資訊
	_, proRes, err := m.ProjectResourceService.GetByListNoPagination(&projectResourceModel.Field{
		ProjectUUID: util.PointerString(input.ProjectUUID),
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// 建立project_resource_id跟resource_name的映射表
	nameToProResIDMap := make(map[string]string)
	for _, res := range proRes {
		nameToProResIDMap[*res.Resources.ResourceName] = *res.ResourceUUID
	}

	// 設置CSV解析的選項
	input.CSVFile.LazyQuotes = true       // 寬鬆地處理引號
	input.CSVFile.TrimLeadingSpace = true // 自動去除每個字段前空格
	input.CSVFile.FieldsPerRecord = -1    // 不強制要求每條記錄擁有相同的字段數

	// 讀取並解析CSV檔案
	records, err := input.CSVFile.ReadAll()
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	taskIdx := [17]int{}
	taskRecordIdx := make(map[string]int)
	var createAllTask []*taskModel.Create
	for i, record := range records {
		if i == 0 {
			// 識別CSV標題行，記錄各欄位的index
			for index, value := range record {
				log.Debug("index: ", index, " value: ", value)
				// 0: gantt project
				if input.FileType == 1 {
					switch value {
					// 根據欄位名稱設置對應的index
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
					case "Coordinator", "協調者":
						taskIdx[7] = index
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
					// 根據欄位名稱設置對應的index
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
					}
				}
			}
			continue
		}

		// 若無大綱編號則略過此記錄
		if record[taskIdx[9]] == "" {
			continue
		}

		// 構建上層大綱編號，用於驗證階層
		outlineNumberSplit := strings.Split(record[taskIdx[9]], ".")
		var topOutlineNumber string
		for i := 0; i < len(outlineNumberSplit)-1; i++ {
			topOutlineNumber += outlineNumberSplit[i]
			if i != len(outlineNumberSplit)-2 {
				topOutlineNumber += "."
			}
		}

		// 驗證檔案中的階層關係
		if _, ok := taskRecordIdx[topOutlineNumber]; topOutlineNumber != "" && !ok {
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, "import errors")
		}

		// 將解析後的資料組合成Create的結構
		createTask := &taskModel.Create{}
		createTask.TaskID = record[taskIdx[0]]
		createTask.TaskName = record[taskIdx[1]]
		if record[taskIdx[2]] != "" {
			// 將日期格式從"/"轉換為"-"
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
			// 將日期格式從"/"轉換為"-"
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

		if record[taskIdx[7]] != "" {
			if nameToProResIDMap[record[taskIdx[7]]] != "" {
				createTask.Coordinator = nameToProResIDMap[record[taskIdx[7]]]
			}
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

				if nameToProResIDMap[name] != "" {
					createTask.Resources = append(createTask.Resources, &resourceModel.TaskSingle{
						ResourceUUID: nameToProResIDMap[name],
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
			// 將日期格式從"/"轉換為"-"
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
			// 將日期格式從"/"轉換為"-"
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

		// 確認無相同outline_number的資料
		if _, ok := taskRecordIdx[record[taskIdx[9]]]; !ok {
			// 紀錄當前task在createAllTask的index位置
			taskRecordIdx[record[taskIdx[9]]] = len(createAllTask)
		}

		createTask.CreatedBy = input.CreatedBy
		createAllTask = assembleToCreateAll(createAllTask, record[taskIdx[9]], "", taskRecordIdx, createTask, input.ProjectUUID)
	}

	httpCode, _ := m.CreateAll(trx, createAllTask)
	if httpCode != code.Successful {
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, "import errors")
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, "Successful import!")
}
