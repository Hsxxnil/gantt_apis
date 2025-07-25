package policy

import (
	policyModel "gantt/internal/interactor/models/policies"
	"gantt/internal/interactor/pkg/util/code"
	"gantt/internal/interactor/pkg/util/log"
	"gantt/internal/router/middleware"

	"github.com/bytedance/sonic"
)

type Manager interface {
	Create(input []*policyModel.PolicyRule) (int, any)
	GetByList() (int, any)
	Delete(input []*policyModel.PolicyRule) (int, any)
}

type manager struct {
}

func Init() Manager {
	return &manager{}
}

func (m *manager) Create(input []*policyModel.PolicyRule) (int, any) {
	var field []*policyModel.PolicyModel
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

	// get all policies
	policies, err := middleware.GetAllPolicies()
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// check if policy already exists
	var addPolices []*policyModel.PolicyModel
	for _, value := range field {
		exists := false
		for _, policy := range policies {
			if value.RoleName == policy[0] && value.Path == policy[1] && value.Method == policy[2] {
				exists = true
				break
			}
		}

		if !exists {
			addPolices = append(addPolices, value)
		}
	}

	result, err := middleware.CreatePolicy(addPolices)
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
	result, err := middleware.GetAllPolicies()
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

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

func (m *manager) Delete(input []*policyModel.PolicyRule) (int, any) {
	var field []*policyModel.PolicyModel
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
