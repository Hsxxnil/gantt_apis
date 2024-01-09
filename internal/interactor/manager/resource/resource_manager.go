package resource

import (
	"errors"
	"github.com/bytedance/sonic"
	userModel "hta/internal/interactor/models/users"
	"hta/internal/interactor/pkg/util"
	userService "hta/internal/interactor/service/user"
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
	UserService     userService.Service
}

func Init(db *gorm.DB) Manager {
	return &manager{
		ResourceService: resourceService.Init(db),
		UserService:     userService.Init(db),
	}
}

func (m *manager) Create(trx *gorm.DB, input *resourceModel.Create) (int, any) {
	defer trx.Rollback()

	// determine if the resource's email is duplicate
	quantity, _ := m.ResourceService.GetByQuantity(&resourceModel.Field{
		Email: util.PointerString(input.Email),
	})

	if quantity > 0 {
		log.Error("Email already exists. Email: ", input.Email)
		return code.BadRequest, code.GetCodeMessage(code.BadRequest, "Email already exists.")
	}

	// transform resource_groups from string slice to string
	resourceGroupByte, err := sonic.Marshal(input.ResourceGroups)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	input.ResourceGroup = string(resourceGroupByte)

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
	resourceByte, err := sonic.Marshal(resourceBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(resourceByte, &output.Resources)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// get all users
	userBase, err := m.UserService.GetByListNoPagination(&userModel.Field{})
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// create a map to store the user's uuid
	userMap := make(map[string]bool)
	for _, user := range userBase {
		if user.ResourceUUID != nil {
			userMap[*user.ResourceUUID] = true
		}
	}
	for i, resource := range output.Resources {
		resource.CreatedBy = *resourceBase[i].CreatedByUsers.Name
		resource.UpdatedBy = *resourceBase[i].UpdatedByUsers.Name

		// transform resource group from string to string slice
		var resourceGroup []string
		err = sonic.Unmarshal([]byte(*resourceBase[i].ResourceGroup), &resourceGroup)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
		resource.ResourceGroups = resourceGroup

		// check if the resource is bind to the user
		if exist := userMap[*resourceBase[i].ResourceUUID]; exist {
			resource.IsBind = true
		} else {
			resource.IsBind = false
		}
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetByListNoPagination(input *resourceModel.Field) (int, any) {
	output := &resourceModel.List{}
	resourceBase, err := m.ResourceService.GetByListNoPagination(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	resourceByte, err := sonic.Marshal(resourceBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(resourceByte, &output.Resources)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// get all users
	userBase, err := m.UserService.GetByListNoPagination(&userModel.Field{})
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// create a map to store the user's uuid
	userMap := make(map[string]bool)
	for _, user := range userBase {
		if user.ResourceUUID != nil {
			userMap[*user.ResourceUUID] = true
		}
	}

	for i, resource := range output.Resources {
		resource.CreatedBy = *resourceBase[i].CreatedByUsers.Name
		resource.UpdatedBy = *resourceBase[i].UpdatedByUsers.Name

		// transform resource group from string to string slice
		var resourceGroup []string
		err = sonic.Unmarshal([]byte(*resourceBase[i].ResourceGroup), &resourceGroup)
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
		resource.ResourceGroups = resourceGroup

		// check if the resource is bind to the user
		if exist := userMap[*resourceBase[i].ResourceUUID]; exist {
			resource.IsBind = true
		} else {
			resource.IsBind = false
		}
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
	resourceByte, _ := sonic.Marshal(resourceBase)
	err = sonic.Unmarshal(resourceByte, &output)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// check if the resource is bind to the user
	quantity, err := m.UserService.GetByQuantity(&userModel.Field{
		ResourceUUID: util.PointerString(input.ResourceUUID),
	})
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	if quantity > 0 {
		output.IsBind = true
	} else {
		output.IsBind = false
	}

	output.CreatedBy = *resourceBase.CreatedByUsers.Name
	output.UpdatedBy = *resourceBase.UpdatedByUsers.Name

	// transform resource group from string to string slice
	var resourceGroup []string
	err = sonic.Unmarshal([]byte(*resourceBase.ResourceGroup), &resourceGroup)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	output.ResourceGroups = resourceGroup

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

	// transform resource_groups from string slice to string
	resourceGroupByte, err := sonic.Marshal(input.ResourceGroups)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	input.ResourceGroup = util.PointerString(string(resourceGroupByte))

	err = m.ResourceService.Update(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, resourceBase.ResourceUUID)
}

func (m *manager) Import(trx *gorm.DB, input *resourceModel.Import) (int, any) {
	defer trx.Rollback()

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

	resourceIdx := [9]int{}
	for i, record := range records {
		if i == 0 {
			for index, value := range record {
				// identify the CSV header row and record the index of each field
				switch value {
				// set the index of each field according to the column name
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

		// determine if the resource is duplicate
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
