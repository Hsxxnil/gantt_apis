package affiliation

import (
	"github.com/bytedance/sonic"
	model "hta/internal/entity/postgresql/db/affiliations"
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

	if input.ID != nil {
		query.Where("id = ?", input.ID)
	}

	if input.UserID != nil {
		query.Where("user_id = ?", input.UserID)
	}

	if input.UserIDs != nil {
		query.Where("user_id in (?)", input.UserIDs)
	}

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

	if input.UserID != nil {
		query.Where("user_id = ?", input.UserID)
	}

	if input.UserIDs != nil {
		query.Where("user_id in (?)", input.UserIDs)
	}

	if input.IsSupervisor != nil {
		query.Where("is_supervisor = ?", input.IsSupervisor)
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

	if input.UserID != nil {
		query.Where("user_id = ?", input.UserID)
	}

	if input.IsSupervisor != nil {
		query.Where("is_supervisor = ?", input.IsSupervisor)
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

	if input.UserID != nil {
		data["user_id"] = input.UserID
	}

	if input.DeptID != nil {
		data["dept_id"] = input.DeptID
	}

	if input.JobTitle != nil {
		data["job_title"] = input.JobTitle
	}

	if input.IsSupervisor != nil {
		data["is_supervisor"] = input.IsSupervisor
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

	if input.UserID != nil {
		query.Where("user_id = ?", input.UserID)
	}

	if input.DeptID != nil {
		query.Where("dept_id = ?", input.DeptID)
	}

	err = query.Delete(&model.Table{}).Error
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}
