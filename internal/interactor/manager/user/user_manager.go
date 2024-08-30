package user

import (
	"errors"
	"fmt"
	affiliationModel "hta/internal/interactor/models/affiliations"
	departmentModel "hta/internal/interactor/models/departments"
	jwxModel "hta/internal/interactor/models/jwx"
	resourceModel "hta/internal/interactor/models/resources"
	"hta/internal/interactor/pkg/email"
	"hta/internal/interactor/pkg/otp"
	"hta/internal/interactor/pkg/util"
	affiliationService "hta/internal/interactor/service/affiliation"
	departmentService "hta/internal/interactor/service/department"
	jwxService "hta/internal/interactor/service/jwx"
	resourceService "hta/internal/interactor/service/resource"

	"github.com/bytedance/sonic"
	"github.com/ggwhite/go-masker"

	userModel "hta/internal/interactor/models/users"
	userService "hta/internal/interactor/service/user"

	"gorm.io/gorm"

	"hta/internal/interactor/pkg/util/code"
	"hta/internal/interactor/pkg/util/log"
)

type Manager interface {
	GetByList(input *userModel.Fields) (int, any)
	GetByListNoPagination(input *userModel.Field) (int, any)
	GetBySingle(input *userModel.Field) (int, any)
	Delete(trx *gorm.DB, input *userModel.Update) (int, any)
	Update(trx *gorm.DB, input *userModel.Update) (int, any)
	Enable(input *userModel.Enable) (int, any)
	ResetPassword(input *userModel.ResetPassword) (int, any)
	Duplicate(input *userModel.Field) (int, any)
	EnableAuthenticator(input *userModel.EnableAuthenticator) (int, any)
	ChangeEmail(input *userModel.ChangeEmail) (int, any)
	VerifyEmail(trx *gorm.DB, input *userModel.VerifyEmail) (int, any)
}

type manager struct {
	UserService        userService.Service
	AffiliationService affiliationService.Service
	DepartmentService  departmentService.Service
	ResourceService    resourceService.Service
	JwxService         jwxService.Service
}

func Init(db *gorm.DB) Manager {
	return &manager{
		UserService:        userService.Init(db),
		AffiliationService: affiliationService.Init(db),
		DepartmentService:  departmentService.Init(db),
		ResourceService:    resourceService.Init(db),
		JwxService:         jwxService.Init(),
	}
}

func (m *manager) GetByList(input *userModel.Fields) (int, any) {
	output := &userModel.List{}
	output.Limit = input.Limit
	output.Page = input.Page

	quantity, userBase, err := m.UserService.GetByList(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output.Total.Total = quantity
	userByte, err := sonic.Marshal(userBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output.Pages = util.Pagination(quantity, output.Limit)
	err = sonic.Unmarshal(userByte, &output.Users)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// collect user IDs for efficient batch processing
	var userIds []*string
	for _, user := range userBase {
		userIds = append(userIds, user.ID)
	}

	// get all job titles
	affiliationBase, err := m.AffiliationService.GetByListNoPagination(&affiliationModel.Field{
		UserIDs: userIds,
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// transform affiliationBase to affiliations.Single
	affiliations := make([]*affiliationModel.Single, len(affiliationBase))
	affiliationByte, _ := sonic.Marshal(affiliationBase)
	err = sonic.Unmarshal(affiliationByte, &affiliations)

	// build maps for efficient lookups and collect department IDs
	affiliationMap := make(map[string][]*affiliationModel.Single)
	var deptIds []*string
	for _, affiliation := range affiliations {
		affiliationMap[affiliation.UserID] = append(affiliationMap[affiliation.UserID], affiliation)
		deptIds = append(deptIds, util.PointerString(affiliation.DeptID))
	}

	// get all departments
	departmentBase, err := m.DepartmentService.GetByListNoPagination(&departmentModel.Field{
		DeptIDs: deptIds,
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// transform departmentBase to departments.Single
	departments := make([]*departmentModel.Single, len(departmentBase))
	departmentByte, _ := sonic.Marshal(departmentBase)
	err = sonic.Unmarshal(departmentByte, &departments)

	// build maps for efficient lookups
	deptMap := make(map[string]*departmentModel.Single)
	for _, dept := range departments {
		deptMap[dept.ID] = dept
	}

	// assign job title and department to each user
	for i, user := range output.Users {
		user.Role = *userBase[i].Roles.DisplayName

		// get the user's job title and department
		if affiliationForUser, ok := affiliationMap[user.ID]; ok {
			for _, affiliation := range affiliationForUser {
				jobTitle := affiliation.JobTitle
				DeptID := affiliation.DeptID
				if dept, ok := deptMap[affiliation.DeptID]; ok {
					deptName := dept.Name
					user.Affiliations = append(user.Affiliations, &affiliationModel.SingleUser{
						JobTitle: jobTitle,
						DeptName: deptName,
						DeptID:   DeptID,
					})
				}
			}
		}
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetByListNoPagination(input *userModel.Field) (int, any) {
	output := &userModel.ListNoPagination{}
	userBase, err := m.UserService.GetByListNoPagination(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	userByte, err := sonic.Marshal(userBase)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	err = sonic.Unmarshal(userByte, &output.Users)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) GetBySingle(input *userModel.Field) (int, any) {
	userBase, err := m.UserService.GetBySingle(input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output := &userModel.Single{}
	userByte, _ := sonic.Marshal(userBase)
	err = sonic.Unmarshal(userByte, &output)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	output.Role = *userBase.Roles.DisplayName

	// get the user's job title
	affiliationBase, err := m.AffiliationService.GetByListNoPagination(&affiliationModel.Field{
		UserID: util.PointerString(input.ID),
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// transform affiliationBase to affiliations.Single
	affiliations := make([]*affiliationModel.Single, len(affiliationBase))
	affiliationByte, _ := sonic.Marshal(affiliationBase)
	err = sonic.Unmarshal(affiliationByte, &affiliations)

	// build maps for efficient lookups and collect department IDs
	affiliationMap := make(map[string][]*affiliationModel.Single)
	var deptIds []*string
	for _, affiliation := range affiliations {
		affiliationMap[affiliation.UserID] = append(affiliationMap[affiliation.UserID], affiliation)
		deptIds = append(deptIds, util.PointerString(affiliation.DeptID))
	}

	// get all departments
	departmentBase, err := m.DepartmentService.GetByListNoPagination(&departmentModel.Field{
		DeptIDs: deptIds,
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// transform departmentBase to departments.Single
	departments := make([]*departmentModel.Single, len(departmentBase))
	departmentByte, _ := sonic.Marshal(departmentBase)
	err = sonic.Unmarshal(departmentByte, &departments)

	// build maps for efficient lookups
	deptMap := make(map[string]*departmentModel.Single)
	for _, dept := range departments {
		deptMap[dept.ID] = dept
	}

	// assign job title and department to the user
	output.Role = *userBase.Roles.DisplayName

	// get the user's job title and department
	if affiliationForUser, ok := affiliationMap[output.ID]; ok {
		for _, affiliation := range affiliationForUser {
			jobTitle := affiliation.JobTitle
			DrptID := affiliation.DeptID
			if dept, ok := deptMap[affiliation.DeptID]; ok {
				deptName := dept.Name
				output.Affiliations = append(output.Affiliations, &affiliationModel.SingleUser{
					JobTitle: jobTitle,
					DeptName: deptName,
					DeptID:   DrptID,
				})
			}
		}
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) Delete(trx *gorm.DB, input *userModel.Update) (int, any) {
	defer trx.Rollback()

	_, err := m.UserService.GetBySingle(&userModel.Field{
		ID: input.ID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.UserService.WithTrx(trx).Delete(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// sync delete affiliation
	err = m.AffiliationService.WithTrx(trx).Delete(&affiliationModel.Field{
		UserID: util.PointerString(input.ID),
	})
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, "Delete ok!")
}

func (m *manager) Update(trx *gorm.DB, input *userModel.Update) (int, any) {
	defer trx.Rollback()

	// validate old password
	if input.Password != nil {
		acknowledge, _, err := m.UserService.AcknowledgeUser(&userModel.Field{
			ID:       input.ID,
			Password: input.OldPassword,
		})
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
		}

		if acknowledge == false {
			return code.PermissionDenied, code.GetCodeMessage(code.PermissionDenied, "Incorrect password.")
		}
	}

	// determine if the username is duplicate
	if input.UserName != nil {
		quantity, _ := m.UserService.GetByQuantity(&userModel.Field{
			UserName: input.UserName,
		})

		if quantity > 0 {
			log.Info("UserName already exists. UserName: ", input.UserName)
			return code.BadRequest, code.GetCodeMessage(code.BadRequest, "User already exists.")
		}
	}

	userBase, err := m.UserService.GetBySingle(&userModel.Field{
		ID: input.ID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// if the user has no resource, bind the resource to the user
	var resourceID *string
	if userBase.ResourceUUID == nil {
		// check if the resource with the same email exists
		resourceBase, err := m.ResourceService.GetBySingle(&resourceModel.Field{
			Email: userBase.Email,
		})
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
		}

		if resourceBase != nil {
			resourceID = resourceBase.ResourceUUID
		} else {
			// sync create resource
			newResourceBase, err := m.ResourceService.WithTrx(trx).Create(&resourceModel.Create{
				ResourceName: *userBase.Name,
				Email:        *userBase.Email,
				CreatedBy:    *input.UpdatedBy,
			})
			if err != nil {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}
			resourceID = newResourceBase.ResourceUUID
		}
		input.ResourceUUID = resourceID
	} else {
		resourceID = userBase.ResourceUUID
	}

	var (
		deptIDs   []*string
		deptNames []*string
	)
	if len(input.Affiliations) > 0 {
		// sync delete affiliation
		err = m.AffiliationService.WithTrx(trx).Delete(&affiliationModel.Field{
			UserID: userBase.ID,
		})
		if err != nil {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}

		// sync create affiliation
		for _, affiliation := range input.Affiliations {
			_, err = m.AffiliationService.WithTrx(trx).Create(&affiliationModel.Create{
				UserID:    *userBase.ID,
				DeptID:    affiliation.DeptID,
				JobTitle:  affiliation.JobTitle,
				CreatedBy: *input.UpdatedBy,
			})
			if err != nil {
				log.Error(err)
				return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
			}

			// collect department IDs for efficient batch processing
			deptIDs = append(deptIDs, util.PointerString(affiliation.DeptID))
		}
	}

	// collect department names for efficient batch processing
	deptBase, err := m.DepartmentService.GetByListNoPagination(&departmentModel.Field{
		DeptIDs: deptIDs,
	})
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	for _, dept := range deptBase {
		deptNames = append(deptNames, dept.Name)
	}

	// transform deptNames to string
	deptByte, err := sonic.Marshal(deptNames)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}
	dept := string(deptByte)

	// sync update resource_group
	err = m.ResourceService.WithTrx(trx).Update(&resourceModel.Update{
		ResourceName:  input.Name,
		ResourceUUID:  *resourceID,
		ResourceGroup: util.PointerString(dept),
		UpdatedBy:     input.UpdatedBy,
	})
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.UserService.WithTrx(trx).Update(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, userBase.ID)
}

func (m *manager) Enable(input *userModel.Enable) (int, any) {
	_, err := m.UserService.GetBySingle(&userModel.Field{
		ID: input.ID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// set default value
	if input.IsEnabled == nil {
		input.IsEnabled = util.PointerBool(true)
	}

	// transform input to Update struct
	update := &userModel.Update{}
	inputByte, err := sonic.Marshal(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(inputByte, &update)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.UserService.Update(update)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, "Enable ok!")
}

func (m *manager) ResetPassword(input *userModel.ResetPassword) (int, any) {
	userBase, err := m.UserService.GetBySingle(&userModel.Field{
		ID: input.ID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// transform input to Update struct
	update := &userModel.Update{}
	inputByte, err := sonic.Marshal(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(inputByte, update)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.UserService.Update(update)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, userBase.ID)
}

func (m *manager) Duplicate(input *userModel.Field) (int, any) {
	output := &userModel.IsDuplicate{}
	input.IsEnabled = util.PointerBool(true)
	quantity, err := m.UserService.GetByQuantity(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	if quantity > 0 {
		log.Info("User already exists. UserName: ", input.FilterUserName, "email: ", input.FilterEmail)
		output.IsDuplicate = true
	} else {
		output.IsDuplicate = false
	}

	return code.Successful, code.GetCodeMessage(code.Successful, output)
}

func (m *manager) EnableAuthenticator(input *userModel.EnableAuthenticator) (int, any) {
	userBase, err := m.UserService.GetBySingle(&userModel.Field{
		ID: input.ID,
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(err)
			return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
		}
	}

	// validate otp
	_, err = otp.ValidateOTP(input.Passcode, *userBase.OtpSecret)
	if err != nil {
		log.Error(err)
		return code.PermissionDenied, code.GetCodeMessage(code.PermissionDenied, "Incorrect passcode.")
	}

	err = m.UserService.Update(&userModel.Update{
		ID:              input.ID,
		IsAuthenticator: util.PointerBool(true),
	})
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	return code.Successful, code.GetCodeMessage(code.Successful, userBase.ID)
}

func (m *manager) ChangeEmail(input *userModel.ChangeEmail) (int, any) {
	userBase, err := m.UserService.GetBySingle(&userModel.Field{
		ID: input.ID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// determine if the email is duplicate
	quantity, _ := m.UserService.GetByQuantity(&userModel.Field{
		UserName: util.PointerString(input.Email),
	})

	if quantity > 0 {
		log.Info("UserName already exists. Email: ", input.Email)
		return code.BadRequest, code.GetCodeMessage(code.BadRequest, "User already exists.")
	}

	// generate access token
	accessToken, err := m.JwxService.CreateAccessToken(&jwxModel.JWX{
		UserID:     userBase.ID,
		Email:      util.PointerString(input.Email),
		Expiration: util.PointerInt64(30),
	})

	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// send link to email
	to := input.Email
	fromAddress := "REMOVED"
	fromName := "PMIS平台"
	mailPwd := "REMOVED"
	subject := "【PMIS平台】請驗證信箱(請勿回覆此郵件)"
	domain := input.Domain
	httpMod := "https"
	// modify localhost port and httpMod for testing
	if input.Domain == "localhost" {
		if input.Port == "" {
			return code.BadRequest, code.GetCodeMessage(code.BadRequest, "Invalid port.")
		}
		domain = fmt.Sprintf("%s:%s", input.Domain, input.Port)
		httpMod = "http"
	}
	verifyLink := fmt.Sprintf("%s://%s/email_verify/%s", httpMod, domain, accessToken.AccessToken)
	message := fmt.Sprintf(`
	<html lang="zh-TW">
		<head>
			<meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
			<style>
				body {
					font-family: 'Arial', sans-serif;
					background-color: #fff;
					color: #000;
				}
		
				.container {
					max-width: 450px;
					margin: 0 auto;
					padding: 20px;
					border-radius: 5px;
					border: 1px solid #cccccc;
					position: relative;
					color: #000;
				}
		
				.header {
					text-align: center;
					color: #000;
				}
		
				.footer {
					text-align: center;
				}
		
				.footerMsg {
					font-size: small;
					color: #737171;
					display: block;
				}
		
				#btnContainer {
					text-align: center;
					margin: 25px;
				}
		
				button {
					background-color: #1f883d;
					border: none;
					border-radius: 5px;
					color: #ffffff;
					padding: 10px 20px;
					text-align: center;
					text-decoration: none;
					display: inline-block;
					font-size: 16px;
					width: 130px;
				}
		
			</style>
		</head>
		<body>
		<div class="header">
			<h2>驗證信箱</h2>
		</div>
		<div class="container">
			<p>親愛的用戶：</p>
			<p>感謝您使用PMIS專案管理平台，請點擊以下按鈕驗證信箱 ：</p>
			<div id="btnContainer">
				<a href="%s">
					<button>
						驗證信箱
					</button>
				</a>
			</div>
			<p>祝您使用愉快！</p>
			<p>此連結時效為30分鐘，若超過時效請重新申請。</p>
			<p class="footerMsg">注意：此郵件由系統自動發出，請勿直接回覆。</p>
		</div>
		</body>
		</html>
		`, verifyLink)

	err = email.SendEmailWithHtml(to, fromAddress, fromName, mailPwd, subject, message)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// mask email
	obscuredEmail := masker.Email(input.Email)

	return code.Successful, code.GetCodeMessage(code.Successful, obscuredEmail)
}

func (m *manager) VerifyEmail(trx *gorm.DB, input *userModel.VerifyEmail) (int, any) {
	defer trx.Rollback()

	userBase, err := m.UserService.GetBySingle(&userModel.Field{
		ID: input.ID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.DoesNotExist, code.GetCodeMessage(code.DoesNotExist, err.Error())
		}

		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// transform input to Update struct
	update := &userModel.Update{}
	inputByte, err := sonic.Marshal(input)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = sonic.Unmarshal(inputByte, &update)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	// sync update the resource's email
	err = m.ResourceService.WithTrx(trx).Update(&resourceModel.Update{
		ResourceUUID: *userBase.ResourceUUID,
		Email:        input.Email,
		UpdatedBy:    input.UpdatedBy,
	})
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	err = m.UserService.WithTrx(trx).Update(update)
	if err != nil {
		log.Error(err)
		return code.InternalServerError, code.GetCodeMessage(code.InternalServerError, err.Error())
	}

	trx.Commit()
	return code.Successful, code.GetCodeMessage(code.Successful, "change email ok!")
}
