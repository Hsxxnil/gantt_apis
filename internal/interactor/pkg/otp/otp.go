package otp

import (
	"errors"
	"github.com/pquerna/otp/totp"
	"hta/internal/interactor/pkg/util/log"
	"time"
)

// GenerateOTP  is used to generate the OTP secret and auth url.
func GenerateOTP(organization, username string) (otpSecret, optAuthUrl string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      organization,
		AccountName: username,
		SecretSize:  15,
	})

	if err != nil {
		log.Error(err)
		return "", "", err
	}

	return key.Secret(), key.URL(), nil
}

// GeneratePasscode is used to generate the OTP code.
func GeneratePasscode(secret string) (passcode string, err error) {
	passcode, err = totp.GenerateCodeCustom(secret, time.Now(), totp.ValidateOpts{
		Period: 60,
	})
	if err != nil {
		log.Error(err)
		return "", err
	}

	return passcode, nil
}

// ValidateOTP is used to validate the OTP code.
func ValidateOTP(passcode, otpSecret string) (otpValid bool, err error) {
	valid := totp.Validate(passcode, otpSecret)
	if !valid {
		log.Error("Passcode is invalid.")
		return false, errors.New("passcode is invalid")
	}

	return true, err
}
