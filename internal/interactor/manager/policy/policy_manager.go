package policy

import (
	"github.com/bytedance/sonic"
	policyModel "hta/internal/interactor/models/policies"
	"hta/internal/interactor/pkg/util/code"
	"hta/internal/interactor/pkg/util/log"
	"hta/internal/router/middleware"
)

type Manager interface {
	Create(input *policyModel.PolicyRule) (int, any)
	GetByList() (int, any)
	Delete(input *policyModel.PolicyRule) (int, any)
}

type manager struct {
}

func Init() Manager {
	return &manager{}
}

func (m *manager) Create(input *policyModel.PolicyRule) (int, any) {
	field := policyModel.PolicyModel{}
	policyByte, err := sonic.Marshal(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(policyByte, &field)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	result, err := middleware.CreatePolicy(field)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	if !result {
		log.Error(err)
		return code.BadRequest, code.GetCodeMessage(code.BadRequest, "Policy already exists.")
	}

	return code.Successful, code.GetCodeMessage(code.Successful, "Create successful!")
}

func (m *manager) GetByList() (int, any) {
	var output []policyModel.Single
	result := middleware.GetAllPolicies()
	if result == nil {
		return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, "Policy does not exist.")
	}

	for _, value := range result {
		output = append(output, policyModel.Single{
			RoleName: value[0],
			Path:     value[1],
			Method:   value[2],
		})
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) Delete(input *policyModel.PolicyRule) (int, any) {
	field := policyModel.PolicyModel{}
	policyByte, err := sonic.Marshal(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(policyByte, &field)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	result, err := middleware.DeletePolicy(field)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	if !result {
		log.Error(err)
		return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, "Policy does not exist.")
	}

	return code.Successful, code.GetCodeMessage(code.Successful, "Delete ok!")
}
