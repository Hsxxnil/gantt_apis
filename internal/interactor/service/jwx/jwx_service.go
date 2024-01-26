package jwx

import (
	"time"

	"hta/config"
	model "hta/internal/interactor/models/jwx"
	"hta/internal/interactor/pkg/jwx"
	"hta/internal/interactor/pkg/util"
	"hta/internal/interactor/pkg/util/log"
)

type Service interface {
	CreateAccessToken(input *model.JWX) (output *model.Token, err error)
	CreateRefreshToken(input *model.JWX) (output *model.Token, err error)
}

type service struct {
}

func Init() Service {
	return &service{}
}

func (s service) CreateAccessToken(input *model.JWX) (output *model.Token, err error) {
	other := map[string]any{
		"user_id":     input.UserID,
		"name":        input.Name,
		"resource_id": input.ResourceID,
		"role":        input.Role,
		"email":       input.Email,
	}

	accessExpiration := util.NowToUTC().Add(time.Minute * 30).Unix()
	if input.Expiration != nil {
		accessExpiration = util.NowToUTC().Add(time.Minute * time.Duration(*input.Expiration)).Unix()
	}

	j := &jwx.JWE{
		PublicKey:     config.AccessPublicKey,
		Other:         other,
		ExpirationKey: accessExpiration,
	}

	j, err = j.Create()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	accessToken := j.Token

	output = &model.Token{
		AccessToken: accessToken,
	}

	return output, nil
}

func (s service) CreateRefreshToken(input *model.JWX) (output *model.Token, err error) {
	other := map[string]any{
		"user_id": input.UserID,
	}

	refreshExpiration := util.NowToUTC().Add(time.Hour * 8).Unix()
	j := &jwx.JWT{
		PrivateKey:    config.RefreshPrivateKey,
		Other:         other,
		ExpirationKey: refreshExpiration,
	}

	j, err = j.Create()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	refreshToken := j.Token
	output = &model.Token{
		RefreshToken: refreshToken,
	}

	return output, nil
}
