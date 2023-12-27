package resource

import (
	"github.com/bytedance/sonic"

	model "hta/internal/entity/postgresql/db/resources"
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

func (s *storage) GetByList(input *model.Base) (quantity int64, output []*model.Table, err error) {
	query := s.db.Model(&model.Table{}).Count(&quantity).Preload(clause.Associations)

	if input.ResourceUUID != nil {
		query.Where("resource_uuid = ?", input.ResourceUUID)
	}

	if input.FilterResourceGroup != nil {
		query.Where("resource_group in (?)", input.FilterResourceGroup)
	}

	if input.Sort.Field != "" && input.Sort.Direction != "" {
		query.Order(input.Sort.Field + " " + input.Sort.Direction)
	}

	// filter
	isFiltered := false
	filter := s.db.Model(&model.Table{})
	if input.FilterResourceName != "" {
		filter.Where("resource_name like ?", "%"+input.FilterResourceName+"%")
		isFiltered = true
	}

	if input.FilterEmail != "" {
		if isFiltered {
			filter.Or("email like ?", "%"+input.FilterEmail+"%")
		} else {
			filter.Where("email like ?", "%"+input.FilterEmail+"%")
		}
	}

	if input.FilterPhone != "" {
		if isFiltered {
			filter.Or("phone like ?", "%"+input.FilterPhone+"%")
		} else {
			filter.Where("phone like ?", "%"+input.FilterPhone+"%")
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
	if input.ResourceUUID != nil {
		query.Where("resource_uuid = ?", input.ResourceUUID)
	}

	if input.ResourceName != nil {
		query.Where("resource_name like ?", "%"+*input.ResourceName+"%")
	}

	err = query.Order("resource_id asc").Find(&output).Error
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return output, nil
}

func (s *storage) GetBySingle(input *model.Base) (output *model.Table, err error) {
	query := s.db.Model(&model.Table{})
	if input.ResourceUUID != nil {
		query.Where("resource_uuid = ?", input.ResourceUUID)
	}

	if input.ResourceName != nil {
		query.Where("resource_name like ?", "%"+*input.ResourceName+"%")
	}

	if input.Email != nil {
		query.Where("email = ?", input.Email)
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
	if input.ResourceUUID != nil {
		query.Where("resource_uuid = ?", input.ResourceUUID)
	}

	if input.ResourceName != nil {
		query.Where("resource_name = ?", input.ResourceName)
	}

	if input.ResourceGroup != nil {
		query.Where("resource_group = ?", input.ResourceGroup)
	}

	if input.Email != nil {
		query.Where("email = ?", input.Email)
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

	if input.ResourceName != nil {
		data["resource_name"] = input.ResourceName
	}

	if input.Email != nil {
		data["email"] = input.Email
	}

	if input.Phone != nil {
		data["phone"] = input.Phone
	}

	if input.StandardCost != nil {
		data["standard_cost"] = input.StandardCost
	}

	if input.TotalCost != nil {
		data["total_cost"] = input.TotalCost
	}

	if input.TotalLoad != nil {
		data["total_load"] = input.TotalLoad
	}

	if input.TotalLoad != nil {
		data["total_load"] = input.TotalLoad
	}

	if input.ResourceGroup != nil {
		data["resource_group"] = input.ResourceGroup
	}

	if input.IsExpand != nil {
		data["is_expand"] = input.IsExpand
	}

	if input.Tags != nil {
		data["tags"] = input.Tags
	}

	if input.UpdatedBy != nil {
		data["updated_by"] = input.UpdatedBy
	}

	if input.ResourceUUID != nil {
		query.Where("resource_uuid = ?", input.ResourceUUID)
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
	if input.ResourceUUID != nil {
		query.Where("resource_uuid = ?", input.ResourceUUID)
	}

	err = query.Delete(&model.Table{}).Error
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}
