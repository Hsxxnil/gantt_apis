package otp

import (
	"errors"
	"github.com/pquerna/otp"
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
		Period:      60,
	})

	if err != nil {
		log.Error(err)
		return "", "", err
	}

	return key.Secret(), key.URL(), nil
}

// GeneratePasscode is used to generate the OTP code.
func GeneratePasscode(secret string) (passcode string, err error) {
	passcode, err = totp.GenerateCodeCustom(secret, time.Now().UTC(), totp.ValidateOpts{
		Period:    60,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		log.Error(err)
		return "", err
	}

	return passcode, nil
}

// ValidateOTP is used to validate the OTP code.
func ValidateOTP(passcode, otpSecret string) (otpValid bool, err error) {
	valid, err := totp.ValidateCustom(passcode, otpSecret, time.Now().UTC(), totp.ValidateOpts{
		Period:    60,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		log.Error(err)
		return false, err
	}

	if !valid {
		log.Error("Passcode is invalid.")
		return false, errors.New("passcode is invalid")
	}

	return true, err
}
