package project_resource

import (
	"errors"
	"gantt/internal/interactor/pkg/util"

	"github.com/bytedance/sonic"

	"gorm.io/gorm"

	projectResourceModel "gantt/internal/interactor/models/project_resources"
	projectResourceService "gantt/internal/interactor/service/project_resource"

	"gantt/internal/interactor/pkg/util/code"
	"gantt/internal/interactor/pkg/util/log"
)

type Manager interface {
	GetByList(input *projectResourceModel.Fields) (int, any)
	GetByProjectList(input *projectResourceModel.ProjectIDs) (int, any)
	GetBySingle(input *projectResourceModel.Field) (int, any)
}

type manager struct {
	ProjectResourceService projectResourceService.Service
}

func Init(db *gorm.DB) Manager {
	return &manager{
		ProjectResourceService: projectResourceService.Init(db),
	}
}

func (m *manager) GetByList(input *projectResourceModel.Fields) (int, any) {
	output := &projectResourceModel.List{}
	output.Limit = input.Limit
	output.Page = input.Page
	quantity, projectResourceBase, err := m.ProjectResourceService.GetByList(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	output.Total.Total = quantity
	output.Pages = util.Pagination(quantity, output.Limit)
	projectResourceByte, err := sonic.Marshal(projectResourceBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(projectResourceByte, &output.ProjectResources)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	for i, proRes := range output.ProjectResources {
		proRes.ResourceID = *projectResourceBase[i].Resources.ResourceID
		proRes.ResourceName = *projectResourceBase[i].Resources.ResourceName
		proRes.Email = *projectResourceBase[i].Resources.Email
		proRes.Phone = *projectResourceBase[i].Resources.Phone
		proRes.StandardCost = *projectResourceBase[i].Resources.StandardCost
		proRes.TotalCost = *projectResourceBase[i].Resources.TotalCost
		proRes.TotalLoad = *projectResourceBase[i].Resources.TotalLoad
		proRes.ResourceGroup = *projectResourceBase[i].Resources.ResourceGroup
		proRes.IsExpand = *projectResourceBase[i].Resources.IsExpand
		proRes.CreatedBy = *projectResourceBase[i].CreatedByUsers.Name
		proRes.UpdatedBy = *projectResourceBase[i].UpdatedByUsers.Name
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetByProjectList(input *projectResourceModel.ProjectIDs) (int, any) {
	var (
		projectResList []projectResourceModel.Single
		output         projectResourceModel.List
	)
	projectResourceBase, err := m.ProjectResourceService.GetByListNoPagination(&projectResourceModel.Field{
		ProjectUUIDs: input.Projects,
	})
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	projectResourceByte, err := sonic.Marshal(projectResourceBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(projectResourceByte, &projectResList)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// filter processed resource
	processedRes := make(map[string]bool)
	for i, proRes := range projectResList {
		resourceID := proRes.ResourceUUID
		if !processedRes[resourceID] {
			proRes.ResourceID = *projectResourceBase[i].Resources.ResourceID
			proRes.ResourceName = *projectResourceBase[i].Resources.ResourceName
			proRes.Email = *projectResourceBase[i].Resources.Email
			proRes.Phone = *projectResourceBase[i].Resources.Phone
			proRes.StandardCost = *projectResourceBase[i].Resources.StandardCost
			proRes.TotalCost = *projectResourceBase[i].Resources.TotalCost
			proRes.TotalLoad = *projectResourceBase[i].Resources.TotalLoad
			proRes.ResourceGroup = *projectResourceBase[i].Resources.ResourceGroup
			proRes.IsExpand = *projectResourceBase[i].Resources.IsExpand
			proRes.CreatedBy = *projectResourceBase[i].CreatedByUsers.Name
			proRes.UpdatedBy = *projectResourceBase[i].UpdatedByUsers.Name
			output.ProjectResources = append(output.ProjectResources, proRes)
			output.Total.Total++
			processedRes[resourceID] = true
		}
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetBySingle(input *projectResourceModel.Field) (int, any) {
	projectResourceBase, err := m.ProjectResourceService.GetBySingle(input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output := &projectResourceModel.Single{}
	projectResourceByte, _ := sonic.Marshal(projectResourceBase)
	err = sonic.Unmarshal(projectResourceByte, &output)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output.ResourceID = *projectResourceBase.Resources.ResourceID
	output.ResourceName = *projectResourceBase.Resources.ResourceName
	output.Email = *projectResourceBase.Resources.Email
	output.Phone = *projectResourceBase.Resources.Phone
	output.StandardCost = *projectResourceBase.Resources.StandardCost
	output.TotalCost = *projectResourceBase.Resources.TotalCost
	output.TotalLoad = *projectResourceBase.Resources.TotalLoad
	output.ResourceGroup = *projectResourceBase.Resources.ResourceGroup
	output.IsExpand = *projectResourceBase.Resources.IsExpand
	output.CreatedBy = *projectResourceBase.CreatedByUsers.Name
	output.UpdatedBy = *projectResourceBase.UpdatedByUsers.Name

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}
