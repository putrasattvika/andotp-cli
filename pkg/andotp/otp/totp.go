package otp

import (
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// generateCodeTOTP generates a token for a TOTP key
func generateCodeTOTP(
	secret string,
	period int,
	digits otp.Digits,
	algorithm otp.Algorithm,
) (string, error) {
	return totp.GenerateCodeCustom(
		secret,
		time.Now(),
		totp.ValidateOpts{
			Period:    uint(period),
			Digits:    digits,
			Algorithm: algorithm,
		},
	)
}
