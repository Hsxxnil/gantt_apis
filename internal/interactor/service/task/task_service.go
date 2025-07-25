package task

import (
	db "gantt/internal/entity/postgresql/db/tasks"
	store "gantt/internal/entity/postgresql/task"
	model "gantt/internal/interactor/models/tasks"
	"gantt/internal/interactor/pkg/util"
	"gantt/internal/interactor/pkg/util/log"
	"gantt/internal/interactor/pkg/util/uuid"

	"github.com/bytedance/sonic"
	"gorm.io/gorm"
)

type Service interface {
	WithTrx(tx *gorm.DB) Service
	Create(input *model.Create) (output *db.Base, err error)
	CreateAll(input []*model.Create) (output []*db.Base, err error)
	GetByList(input *model.Field) (quantity int64, output []*db.Base, err error)
	GetByListNoPagination(input *model.Field) (output []*db.Base, err error)
	GetByListNoQuantity(input *model.Field) (output []*db.Base, err error)
	GetBySingle(input *model.Field) (output *db.Base, err error)
	GetByQuantity(input *model.Field) (quantity int64, err error)
	Update(input *model.Update) (err error)
	Delete(input *model.Field) (err error)
	GetByLastTaskID(input *model.Field) (output *db.Base, err error)
	GetByLastOutlineNumber(input *model.Field) (output *db.Base, err error)
	GetByMinStartMaxEnd(input *model.Field) (output []*db.Base, err error)
}

type service struct {
	Repository store.Entity
}

func Init(db *gorm.DB) Service {
	return &service{
		Repository: store.Init(db),
	}
}

func (s *service) WithTrx(tx *gorm.DB) Service {
	return &service{
		Repository: s.Repository.WithTrx(tx),
	}
}

func (s *service) Create(input *model.Create) (output *db.Base, err error) {
	base := &db.Base{}
	marshal, err := sonic.Marshal(input)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = sonic.Unmarshal(marshal, &base)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	base.TaskUUID = util.PointerString(uuid.CreatedUUIDString())
	base.CreatedAt = util.PointerTime(util.NowToUTC())
	base.UpdatedAt = util.PointerTime(util.NowToUTC())
	base.UpdatedBy = util.PointerString(input.CreatedBy)
	err = s.Repository.Create(base)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	marshal, err = sonic.Marshal(base)
	if err != nil {
		log.Error(err)

		return nil, err
	}

	err = sonic.Unmarshal(marshal, &output)
	if err != nil {
		log.Error(err)

		return nil, err
	}

	return output, nil
}

func (s *service) CreateAll(input []*model.Create) (output []*db.Base, err error) {
	var base []*db.Base
	marshal, err := sonic.Marshal(input)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = sonic.Unmarshal(marshal, &base)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	for i, field := range base {
		field.TaskUUID = util.PointerString(uuid.CreatedUUIDString())
		field.CreatedAt = util.PointerTime(util.NowToUTC())
		field.UpdatedAt = util.PointerTime(util.NowToUTC())
		field.UpdatedBy = util.PointerString(input[i].CreatedBy)
	}

	err = s.Repository.CreateAll(base)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	marshal, err = sonic.Marshal(base)
	if err != nil {
		log.Error(err)

		return nil, err
	}

	err = sonic.Unmarshal(marshal, &output)
	if err != nil {
		log.Error(err)

		return nil, err
	}

	return output, nil
}

func (s *service) GetByList(input *model.Field) (quantity int64, output []*db.Base, err error) {
	field := &db.Base{}
	marshal, err := sonic.Marshal(input)
	if err != nil {
		log.Error(err)
		return 0, nil, err
	}

	err = sonic.Unmarshal(marshal, &field)
	if err != nil {
		log.Error(err)
		return 0, nil, err
	}

	quantity, fields, err := s.Repository.GetByList(field)
	if err != nil {
		log.Error(err)
		return 0, output, err
	}

	marshal, err = sonic.Marshal(fields)
	if err != nil {
		log.Error(err)
		return 0, nil, err
	}

	err = sonic.Unmarshal(marshal, &output)
	if err != nil {
		log.Error(err)
		return 0, nil, err
	}

	return quantity, output, nil
}

func (s *service) GetByListNoPagination(input *model.Field) (output []*db.Base, err error) {
	field := &db.Base{}
	marshal, err := sonic.Marshal(input)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = sonic.Unmarshal(marshal, &field)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	fields, err := s.Repository.GetByListNoPagination(field)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	marshal, err = sonic.Marshal(fields)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = sonic.Unmarshal(marshal, &output)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return output, nil
}

func (s *service) GetByListNoQuantity(input *model.Field) (output []*db.Base, err error) {
	field := &db.Base{}
	marshal, err := sonic.Marshal(input)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = sonic.Unmarshal(marshal, &field)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	fields, err := s.Repository.GetByListNoQuantity(field)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	marshal, err = sonic.Marshal(fields)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = sonic.Unmarshal(marshal, &output)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return output, nil
}

func (s *service) GetBySingle(input *model.Field) (output *db.Base, err error) {
	field := &db.Base{}
	marshal, err := sonic.Marshal(input)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = sonic.Unmarshal(marshal, &field)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	single, err := s.Repository.GetBySingle(field)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	marshal, err = sonic.Marshal(single)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = sonic.Unmarshal(marshal, &output)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return output, nil
}

func (s *service) Delete(input *model.Field) (err error) {
	field := &db.Base{}
	marshal, err := sonic.Marshal(input)
	if err != nil {
		log.Error(err)
		return err
	}

	err = sonic.Unmarshal(marshal, &field)
	if err != nil {
		log.Error(err)
		return err
	}

	err = s.Repository.Delete(field)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (s *service) Update(input *model.Update) (err error) {
	field := &db.Base{}
	marshal, err := sonic.Marshal(input)
	if err != nil {
		log.Error(err)
		return err
	}

	err = sonic.Unmarshal(marshal, &field)
	if err != nil {
		log.Error(err)
		return err
	}

	err = s.Repository.Update(field)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (s *service) GetByQuantity(input *model.Field) (quantity int64, err error) {
	field := &db.Base{}
	marshal, err := sonic.Marshal(input)
	if err != nil {
		log.Error(err)
		return 0, err
	}

	err = sonic.Unmarshal(marshal, &field)
	if err != nil {
		log.Error(err)
		return 0, err
	}

	quantity, err = s.Repository.GetByQuantity(field)
	if err != nil {
		log.Error(err)
		return 0, err
	}

	return quantity, nil
}

func (s *service) GetByLastTaskID(input *model.Field) (output *db.Base, err error) {
	field := &db.Base{}
	marshal, err := sonic.Marshal(input)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = sonic.Unmarshal(marshal, &field)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	single, err := s.Repository.GetByLastTaskID(field)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	marshal, err = sonic.Marshal(single)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = sonic.Unmarshal(marshal, &output)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return output, nil
}

func (s *service) GetByLastOutlineNumber(input *model.Field) (output *db.Base, err error) {
	field := &db.Base{}
	marshal, err := sonic.Marshal(input)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = sonic.Unmarshal(marshal, &field)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	single, err := s.Repository.GetByLastOutlineNumber(field)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	marshal, err = sonic.Marshal(single)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = sonic.Unmarshal(marshal, &output)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return output, nil
}

func (s *service) GetByMinStartMaxEnd(input *model.Field) (output []*db.Base, err error) {
	field := &db.Base{}
	marshal, err := sonic.Marshal(input)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = sonic.Unmarshal(marshal, &field)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	fields, err := s.Repository.GetByMinStartMaxEnd(field)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	marshal, err = sonic.Marshal(fields)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = sonic.Unmarshal(marshal, &output)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return output, nil
}
