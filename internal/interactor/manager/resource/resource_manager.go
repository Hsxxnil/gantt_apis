package resource

import (
	"encoding/json"
	"errors"
	"hta/internal/interactor/pkg/util"
	"strconv"

	"gorm.io/gorm"

	resourceModel "hta/internal/interactor/models/resources"
	resourceService "hta/internal/interactor/service/resource"

	"hta/internal/interactor/pkg/util/code"
	"hta/internal/interactor/pkg/util/log"
)

type Manager interface {
	Create(trx *gorm.DB, input *resourceModel.Create) (int, any)
	GetByList(input *resourceModel.Fields) (int, any)
	GetByListNoPagination(input *resourceModel.Field) (int, any)
	GetBySingle(input *resourceModel.Field) (int, any)
	Delete(input *resourceModel.Field) (int, any)
	Update(input *resourceModel.Update) (int, any)
	Import(trx *gorm.DB, input *resourceModel.Import) (int, any)
}

type manager struct {
	ResourceService resourceService.Service
}

func Init(db *gorm.DB) Manager {
	return &manager{
		ResourceService: resourceService.Init(db),
	}
}

func (m *manager) Create(trx *gorm.DB, input *resourceModel.Create) (int, any) {
	defer trx.Rollback()

	resourceBase, err := m.ResourceService.WithTrx(trx).Create(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, resourceBase.ResourceUUID)
}

func (m *manager) GetByList(input *resourceModel.Fields) (int, any) {
	output := &resourceModel.List{}
	output.Limit = input.Limit
	output.Page = input.Page
	quantity, resourceBase, err := m.ResourceService.GetByList(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	output.Total.Total = quantity
	output.Pages = util.Pagination(quantity, output.Limit)
	resourceByte, err := json.Marshal(resourceBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = json.Unmarshal(resourceByte, &output.Resources)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	for i, resource := range output.Resources {
		resource.CreatedBy = *resourceBase[i].CreatedByUsers.Name
		resource.UpdatedBy = *resourceBase[i].UpdatedByUsers.Name
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetByListNoPagination(input *resourceModel.Field) (int, any) {
	output := &resourceModel.List{}
	quantity, resourceBase, err := m.ResourceService.GetByListNoPagination(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	output.Total.Total = quantity
	resourceByte, err := json.Marshal(resourceBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = json.Unmarshal(resourceByte, &output.Resources)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	for i, resource := range output.Resources {
		resource.CreatedBy = *resourceBase[i].CreatedByUsers.Name
		resource.UpdatedBy = *resourceBase[i].UpdatedByUsers.Name
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetBySingle(input *resourceModel.Field) (int, any) {
	resourceBase, err := m.ResourceService.GetBySingle(input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output := &resourceModel.Single{}
	resourceByte, _ := json.Marshal(resourceBase)
	err = json.Unmarshal(resourceByte, &output)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output.CreatedBy = *resourceBase.CreatedByUsers.Name
	output.UpdatedBy = *resourceBase.UpdatedByUsers.Name

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) Delete(input *resourceModel.Field) (int, any) {
	_, err := m.ResourceService.GetBySingle(&resourceModel.Field{
		ResourceUUID: input.ResourceUUID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.ResourceService.Delete(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, "Delete ok!")
}

func (m *manager) Update(input *resourceModel.Update) (int, any) {
	resourceBase, err := m.ResourceService.GetBySingle(&resourceModel.Field{
		ResourceUUID: input.ResourceUUID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.ResourceService.Update(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, resourceBase.ResourceUUID)
}

func (m *manager) Import(trx *gorm.DB, input *resourceModel.Import) (int, any) {
	defer trx.Rollback()

	//設置CSV解析的選項
	input.CSVFile.LazyQuotes = true       //寬鬆地處理引號
	input.CSVFile.TrimLeadingSpace = true //自動去除每個字段前空格
	input.CSVFile.FieldsPerRecord = -1    //不強制要求每條記錄擁有相同的字段數

	//讀取並解析CSV檔案
	records, err := input.CSVFile.ReadAll()
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	resourceIdx := [9]int{}
	for i, record := range records {
		if i == 0 {
			for index, value := range record {
				//識別CSV標題行，記錄各欄位的index
				switch value {
				//根據欄位名稱設置對應的index
				//case "ID", "編號":
				//	resourceIdx[0] = index
				case "Name", "姓名":
					resourceIdx[1] = index
				//case "Default role", "預設角色":
				//	resourceIdx[2] = index
				case "e-mail", "E-mail":
					resourceIdx[3] = index
				case "Phone", "電話":
					resourceIdx[4] = index
				case "Standard rate":
					resourceIdx[5] = index
				case "Total cost":
					resourceIdx[6] = index
				case "Total load":
					resourceIdx[7] = index
				case "Group":
					resourceIdx[8] = index
				}
			}
			continue
		}

		standardCost, err := strconv.ParseFloat(record[resourceIdx[5]], 64)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}

		totalCost, err := strconv.ParseFloat(record[resourceIdx[6]], 64)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}

		totalLoad, err := strconv.ParseFloat(record[resourceIdx[7]], 64)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}

		// 判斷資源是否重複
		quantity, _ := m.ResourceService.GetByQuantity(&resourceModel.Field{
			ResourceName:  util.PointerString(record[resourceIdx[1]]),
			ResourceGroup: util.PointerString(record[resourceIdx[8]]),
		})
		if quantity > 0 {
			continue
		}

		_, err = m.ResourceService.WithTrx(trx).Create(&resourceModel.Create{
			ResourceName:  record[resourceIdx[1]],
			Email:         record[resourceIdx[3]],
			Phone:         record[resourceIdx[4]],
			StandardCost:  standardCost,
			TotalCost:     totalCost,
			TotalLoad:     totalLoad,
			ResourceGroup: record[resourceIdx[8]],
			CreatedBy:     input.CreatedBy,
		})

		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, "Successful import!")
}
