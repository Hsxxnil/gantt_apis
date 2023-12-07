package task

import (
	"encoding/json"
	model "hta/internal/entity/postgresql/db/tasks"
	"hta/internal/interactor/pkg/util/log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Entity interface {
	WithTrx(trx *gorm.DB) Entity
	Create(input *model.Base) (err error)
	CreateAll(input []*model.Base) (err error)
	GetByList(input *model.Base) (quantity int64, output []*model.Table, err error)
	GetByListNoPagination(input *model.Base) (output []*model.Table, err error)
	GetByListNoQuantity(input *model.Base) (output []*model.Table, err error)
	GetBySingle(input *model.Base) (output *model.Table, err error)
	GetByQuantity(input *model.Base) (quantity int64, err error)
	GetByLastTaskID(input *model.Base) (output *model.Table, err error)
	GetByLastOutlineNumber(input *model.Base) (output *model.Table, err error)
	GetByMinStartMaxEnd(input *model.Base) (output []*model.Table, err error)
	Delete(input *model.Base) (err error)
	Update(input *model.Base) (err error)
}

type storage struct {
	db *gorm.DB
}

func Init(db *gorm.DB) Entity {
	return &storage{
		db: db,
	}
}

func (s *storage) WithTrx(trx *gorm.DB) Entity {
	return &storage{
		db: trx,
	}
}

func (s *storage) Create(input *model.Base) (err error) {
	marshal, err := json.Marshal(input)
	if err != nil {
		log.Error(err)
		return err
	}

	data := &model.Table{}
	err = json.Unmarshal(marshal, data)
	if err != nil {
		log.Error(err)
		return err
	}

	err = s.db.Model(&model.Table{}).Omit(clause.Associations).Create(&data).Error
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (s *storage) CreateAll(input []*model.Base) (err error) {
	marshal, err := json.Marshal(input)
	if err != nil {
		log.Error(err)
		return err
	}

	data := &[]model.Table{}
	err = json.Unmarshal(marshal, data)
	if err != nil {
		log.Error(err)
		return err
	}

	err = s.db.Model(&model.Table{}).Omit(clause.Associations).Create(&data).Error
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (s *storage) GetByList(input *model.Base) (quantity int64, output []*model.Table, err error) {
	query := s.db.Model(&model.Table{}).Count(&quantity)

	if input.TaskUUID != nil {
		query.Where("task_uuid = ?", input.TaskUUID)
	}

	if input.DeletedTaskUUIDs != nil {
		query.Where("task_uuid in (?)", input.DeletedTaskUUIDs)
	}

	err = query.Count(&quantity).Order("created_at desc").Find(&output).Error
	if err != nil {
		log.Error(err)
		return 0, nil, err
	}

	return quantity, output, nil
}

func (s *storage) GetByListNoPagination(input *model.Base) (output []*model.Table, err error) {
	query := s.db.Model(&model.Table{}).Preload("TaskResources.Resources.Resources").Preload(clause.Associations)
	if input.TaskUUID != nil {
		query.Where("task_uuid = ?", input.TaskUUID)
	}

	if input.ProjectUUID != nil {
		query.Where("project_uuid = ?", input.ProjectUUID)
	}

	// filter
	//isFiltered := false
	filter := s.db.Model(&model.Table{})
	if input.FilterMilestone {
		filter.Where("duration = 0 and date(start_date) = date(end_date)")
		//isFiltered = true
	}

	query.Where(filter)

	err = query.Order(`(select array_agg(nullif(part, ''):: int) from unnest(string_to_array(outline_number, '.')) as part) asc`).Find(&output).Error
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return output, nil
}

func (s *storage) GetByListNoQuantity(input *model.Base) (output []*model.Table, err error) {
	query := s.db.Model(&model.Table{})
	if input.TaskUUID != nil {
		query.Where("task_uuid = ?", input.TaskUUID)
	}

	if input.ProjectUUID != nil {
		query.Where("project_uuid = ?", input.ProjectUUID)
	}

	err = query.Order("created_at desc").Find(&output).Error
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return output, nil
}

func (s *storage) GetBySingle(input *model.Base) (output *model.Table, err error) {
	query := s.db.Model(&model.Table{}).Preload("TaskResources.Resources.Resources").Preload(clause.Associations)
	if input.TaskUUID != nil {
		query.Where("task_uuid = ?", input.TaskUUID)
	}

	err = query.First(&output).Error
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return output, nil
}

func (s *storage) GetByQuantity(input *model.Base) (quantity int64, err error) {
	query := s.db.Model(&model.Table{})
	if input.TaskUUID != nil {
		query.Where("task_uuid = ?", input.TaskUUID)
	}

	if input.ProjectUUID != nil {
		query.Where("project_uuid = ?", input.ProjectUUID)
	}

	if input.OutlineNumber != nil {
		query.Where("outline_number LIKE ?", *input.OutlineNumber+".%")
	}

	err = query.Count(&quantity).Select("*").Error
	if err != nil {
		log.Error(err)
		return 0, err
	}

	return quantity, nil
}

func (s *storage) GetByLastTaskID(input *model.Base) (output *model.Table, err error) {
	query := s.db.Model(&model.Table{}).Preload(clause.Associations)
	if input.TaskUUID != nil {
		query.Where("task_uuid = ?", input.TaskUUID)
	}

	err = query.Order("task_id desc").First(&output).Error
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return output, nil
}

func (s *storage) GetByLastOutlineNumber(input *model.Base) (output *model.Table, err error) {
	query := s.db.Model(&model.Table{}).Preload(clause.Associations)
	if input.OutlineNumber != nil {
		query.Where("outline_number LIKE ?", *input.OutlineNumber+".%")
	}

	if input.ProjectUUID != nil {
		query.Where("project_uuid = ?", input.ProjectUUID)
	}

	// sorting in descending order in the case of infinite levels
	err = query.
		Order(`(select array_agg(nullif(part, ''):: int) from unnest(string_to_array(outline_number, '.')) as part) desc`).
		First(&output).Error
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return output, nil
}

func (s *storage) GetByMinStartMaxEnd(input *model.Base) (output []*model.Table, err error) {
	query := s.db.Model(&model.Table{})
	subQuery1 := s.db.Model(&model.Table{}).Select("min(baseline_start_date)")
	subQuery2 := s.db.Model(&model.Table{}).Select("max(baseline_end_date)")

	if input.TaskUUID != nil {
		query.Where("task_uuid = ?", input.TaskUUID)
	}

	if input.ProjectUUID != nil {
		query.Where("project_uuid = ?", input.ProjectUUID)
		subQuery1.Where("project_uuid = ?", input.ProjectUUID)
		subQuery2.Where("project_uuid = ?", input.ProjectUUID)
	}

	if input.DeletedTaskUUIDs != nil {
		subQuery1.Where("task_uuid not in (?)", input.DeletedTaskUUIDs)
		subQuery2.Where("task_uuid not in (?)", input.DeletedTaskUUIDs)
	}

	err = query.Where("baseline_start_date = (?)", subQuery1).Or("baseline_end_date =(?)", subQuery2).Find(&output).Error
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return output, nil
}

func (s *storage) Update(input *model.Base) (err error) {
	query := s.db.Model(&model.Table{}).Omit(clause.Associations)
	data := map[string]any{}

	if input.TaskName != nil {
		data["task_name"] = input.TaskName
	}

	if input.StartDate != nil {
		data["start_date"] = input.StartDate
	}

	if input.EndDate != nil {
		data["end_date"] = input.EndDate
	}

	if input.BaselineStartDate != nil {
		data["baseline_start_date"] = input.BaselineStartDate
	}

	if input.BaselineEndDate != nil {
		data["baseline_end_date"] = input.BaselineEndDate
	}

	if input.Duration != nil {
		data["duration"] = input.Duration
	}

	if input.Progress != nil {
		data["progress"] = input.Progress
	}

	if input.Cost != nil {
		data["cost"] = input.Cost
	}

	if input.Coordinator != nil {
		data["coordinator"] = input.Coordinator
	} else {
		data["coordinator"] = nil
	}

	if input.Segment != nil {
		data["segment"] = input.Segment
	} else {
		data["segment"] = nil
	}

	if input.Indicator != nil {
		data["indicator"] = input.Indicator
	} else {
		data["indicator"] = nil
	}

	if input.Predecessor != nil {
		data["predecessor"] = input.Predecessor
	}

	if input.OutlineNumber != nil {
		data["outline_number"] = input.OutlineNumber
	}

	if input.Assignments != nil {
		data["assignments"] = input.Assignments
	}

	if input.TaskColor != nil {
		data["task_color"] = input.TaskColor
	}

	if input.WebLink != nil {
		data["web_link"] = input.WebLink
	}

	if input.IsSubTask != nil {
		data["is_subtask"] = input.IsSubTask
	}

	if input.ProjectUUID != nil {
		data["project_uuid"] = input.ProjectUUID
	}

	if input.Notes != nil {
		data["notes"] = input.Notes
	}

	if input.UpdatedBy != nil {
		data["updated_by"] = input.UpdatedBy
	}

	if input.TaskUUID != nil {
		query.Where("task_uuid = ?", input.TaskUUID)
	}

	err = query.Select("*").Updates(data).Error
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (s *storage) Delete(input *model.Base) (err error) {
	query := s.db.Model(&model.Table{}).Omit(clause.Associations)
	if input.TaskUUID != nil {
		query.Where("task_uuid = ?", input.TaskUUID)
	}

	if input.DeletedTaskUUIDs != nil {
		query.Where("task_uuid in (?)", input.DeletedTaskUUIDs)
	}

	if input.ProjectUUID != nil {
		query.Where("project_uuid = ?", input.ProjectUUID)
	}

	err = query.Delete(&model.Table{}).Error
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}
