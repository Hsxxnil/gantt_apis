package project_resource

import (
	"github.com/bytedance/sonic"
	model "hta/internal/entity/postgresql/db/project_resources"
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
	marshal, err := sonic.Marshal(input)
	if err != nil {
		log.Error(err)
		return err
	}

	data := &model.Table{}
	err = sonic.Unmarshal(marshal, data)
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
	marshal, err := sonic.Marshal(input)
	if err != nil {
		log.Error(err)
		return err
	}

	data := &[]model.Table{}
	err = sonic.Unmarshal(marshal, data)
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
	query := s.db.Model(&model.Table{}).Count(&quantity).
		Joins("Resources").Preload(clause.Associations)

	if input.ID != nil {
		query.Where("id = ?", input.ID)
	}

	// filter
	isFiltered := false
	filter := s.db.Model(&model.Table{})
	if input.FilterRole != "" {
		filter.Where("project_resources.role like ?", "%"+input.FilterRole+"%")
		isFiltered = true
	}

	if input.FilterResourceName != "" {
		if isFiltered {
			filter.Or(`"Resources".resource_name like ?`, "%"+input.FilterResourceName+"%")
		} else {
			filter.Where(`"Resources".resource_name like ?`, "%"+input.FilterResourceName+"%")
		}
	}

	if input.FilterResourceGroup != "" {
		if isFiltered {
			filter.Or(`"Resources".resource_group like ?`, "%"+input.FilterResourceGroup+"%")
		} else {
			filter.Where(`"Resources".resource_group like ?`, "%"+input.FilterResourceGroup+"%")
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

	if input.ID != nil {
		query.Where("id = ?", input.ID)
	}

	if input.ProjectUUID != nil {
		query.Where("project_uuid = ?", input.ProjectUUID)
	}

	if input.ProjectUUIDs != nil {
		query.Where("project_uuid in (?)", input.ProjectUUIDs)
	}

	if input.Role != nil {
		query.Where("role = ?", input.Role)
	}

	if input.ResourceUUID != nil {
		query.Where("resource_uuid = ?", input.ResourceUUID)
	}

	if input.ResourceUUIDs != nil {
		query.Where("resource_uuid in (?)", input.ResourceUUIDs)
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
	if input.ID != nil {
		query.Where("id = ?", input.ID)
	}

	if input.Role != nil {
		query.Where("role = ?", input.Role)
	}

	if input.ProjectUUID != nil {
		query.Where("project_uuid = ?", input.ProjectUUID)
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
	if input.ID != nil {
		query.Where("id = ?", input.ID)
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

	if input.Role != nil {
		data["role"] = input.Role
	}

	if input.IsEditable != nil {
		data["is_editable"] = input.IsEditable
	}

	if input.UpdatedBy != nil {
		data["updated_by"] = input.UpdatedBy
	}

	if input.ID != nil {
		query.Where("id = ?", input.ID)
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
	if input.ID != nil {
		query.Where("id = ?", input.ID)
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
