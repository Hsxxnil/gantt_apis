package project

import (
	"encoding/json"
	model "hta/internal/entity/postgresql/db/projects"
	"hta/internal/interactor/pkg/util/log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Entity interface {
	WithTrx(trx *gorm.DB) Entity
	Create(input *model.Base) (err error)
	GetByList(input *model.Base) (quantity int64, output []*model.Table, err error)
	GetByListNoPagination(input *model.Base) (output []*model.Table, err error)
	GetBySingle(input *model.Base) (output *model.Table, err error)
	GetByQuantity(input *model.Base) (quantity int64, err error)
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

func (s *storage) GetByList(input *model.Base) (quantity int64, output []*model.Table, err error) {
	query := s.db.Model(&model.Table{}).Count(&quantity).Preload(clause.Associations)

	if input.ProjectUUID != nil {
		query.Where("project_uuid = ?", input.ProjectUUID)
	}

	if input.CreatedBy != nil {
		query.Where("created_by = ?", input.CreatedBy)
	}

	if input.FilterStatus != nil {
		query.Where("status in (?)", input.FilterStatus)
	}

	if input.FilterTypes != nil {
		query.Where("type in (?)", input.FilterTypes)
	}

	// filter
	isFiltered := false
	filter := s.db.Model(&model.Table{})
	if input.FilterClient != "" {
		filter.Where("client like ?", "%"+input.FilterClient+"%")
		isFiltered = true
	}

	if input.FilterName != "" {
		if isFiltered {
			filter.Or("project_name like ?", "%"+input.FilterName+"%")
		} else {
			filter.Where("project_name like ?", "%"+input.FilterName+"%")
		}
	}

	if input.FilterManagers != nil {
		if isFiltered {
			filter.Or("manager in (?)", input.FilterManagers)
		} else {
			filter.Where("manager in (?)", input.FilterManagers)
		}
	}

	if input.FilterCode != "" {
		if isFiltered {
			filter.Or("code like ?", "%"+input.FilterCode+"%")
		} else {
			filter.Where("code like ?", "%"+input.FilterCode+"%")
		}
	}

	if input.FilterStartDate != nil && input.FilterEndDate == nil {
		if isFiltered {
			filter.Or("start_date >= ?", input.FilterStartDate)
		} else {
			filter.Where("start_date >= ?", input.FilterStartDate)
		}
	}

	if input.FilterEndDate != nil && input.FilterStartDate == nil {
		if isFiltered {
			filter.Or("end_date <= ?", input.FilterEndDate)
		} else {
			filter.Where("end_date <= ?", input.FilterEndDate)
		}
	}

	if input.FilterStartDate != nil && input.FilterEndDate != nil {
		if isFiltered {
			filter.Or("start_date >= ? and end_date <= ?", input.FilterStartDate, input.FilterEndDate)
		} else {
			filter.Where("start_date >= ? and end_date <= ?", input.FilterStartDate, input.FilterEndDate)
		}
	}

	query.Where(filter)

	err = query.Count(&quantity).Offset(int((input.Page - 1) * input.Limit)).
		Limit(int(input.Limit)).Order("created_at desc").Find(&output).Error
	if err != nil {
		log.Error(err)
		return 0, nil, err
	}

	return quantity, output, nil
}

func (s *storage) GetByListNoPagination(input *model.Base) (output []*model.Table, err error) {
	query := s.db.Model(&model.Table{}).Preload(clause.Associations)

	if input.ProjectUUID != nil {
		query.Where("project_uuid = ?", input.ProjectUUID)
	}

	if input.CreatedBy != nil {
		query.Where("created_by = ?", input.CreatedBy)
	}

	if input.ProjectIDs != nil {
		query.Where("project_uuid in (?)", input.ProjectIDs)
	}

	err = query.Order("created_at desc").Find(&output).Error
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return output, nil
}

func (s *storage) GetBySingle(input *model.Base) (output *model.Table, err error) {
	query := s.db.Model(&model.Table{}).Preload(clause.Associations)
	if input.ProjectUUID != nil {
		query.Where("project_uuid = ?", input.ProjectUUID)
	}

	if input.CreatedBy != nil {
		query.Where("created_by = ?", input.CreatedBy)
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
	if input.ProjectUUID != nil {
		query.Where("project_uuid = ?", input.ProjectUUID)
	}

	if input.CreatedBy != nil {
		query.Where("created_by = ?", input.CreatedBy)
	}

	err = query.Count(&quantity).Select("*").Error
	if err != nil {
		log.Error(err)
		return 0, err
	}

	return quantity, nil
}

func (s *storage) Update(input *model.Base) (err error) {
	query := s.db.Model(&model.Table{}).Omit(clause.Associations)
	data := map[string]any{}

	if input.ProjectName != nil {
		data["project_name"] = input.ProjectName
	}

	if input.ProjectName != nil {
		data["project_name"] = input.ProjectName
	}

	if input.Type != nil {
		data["type"] = input.Type
	}

	if input.Code != nil {
		data["code"] = input.Code
	}

	if input.Manager != nil {
		data["manager"] = input.Manager
	}

	if input.StartDate != nil {
		data["start_date"] = input.StartDate
	}

	if input.EndDate != nil {
		data["end_date"] = input.EndDate
	}

	if input.Client != nil {
		data["client"] = input.Client
	}

	if input.Status != nil {
		data["status"] = input.Status
	}

	if input.UpdatedBy != nil {
		data["updated_by"] = input.UpdatedBy
	}

	if input.ProjectUUID != nil {
		query.Where("project_uuid = ?", input.ProjectUUID)
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
