package otp

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/awnumar/memguard"
	memguardcore "github.com/awnumar/memguard/core"
	"github.com/pquerna/otp"
)

type otpGenerateCodeFunc func(secret string, period int, digits otp.Digits, algorithm otp.Algorithm) (string, error)

var otpTypeMapping = map[string]otpGenerateCodeFunc{
	"TOTP": generateCodeTOTP,
}

var otpAlgorithmMapping = map[string]otp.Algorithm{
	"SHA1":   otp.AlgorithmSHA1,
	"SHA256": otp.AlgorithmSHA256,
	"SHA512": otp.AlgorithmSHA512,
	"MD5":    otp.AlgorithmMD5,
}

// OTPKey holds the structure of a single OTP key from backup
type OTPKey struct {
	Issuer       string `json:"issuer"`
	Label        string `json:"label"`
	Digits       otp.Digits
	DigitsInt    int    `json:"digits"`
	OTPType      string `json:"type"`
	Algorithm    otp.Algorithm
	AlgorithmStr string   `json:"algorithm"`
	Period       int      `json:"period"`
	Tags         []string `json:"tags"`

	// Will always be empty if OTPKeysFromJSON() was used
	Secret string `json:"secret"`

	secretEnclave *memguard.Enclave
}

// OTPKeysFromJSON parses OTP keys from andOTP JSON backup
func OTPKeysFromJSON(otpJSON []byte) ([]*OTPKey, error) {
	otpKeys := make([]*OTPKey, 0)

	if err := json.Unmarshal(otpJSON, &otpKeys); err != nil {
		return nil, err
	}

	for _, otpKey := range otpKeys {
		// Create a new memguard enclave for the OTP secret and wipe the
		// plaintext secret from memory
		secretBytes := []byte(otpKey.Secret)

		otpKey.secretEnclave = memguard.NewEnclave(secretBytes)
		memguardcore.Wipe(secretBytes)
		otpKey.Secret = ""

		// Validations & parsing
		if _, ok := otpAlgorithmMapping[otpKey.AlgorithmStr]; !ok {
			return nil, fmt.Errorf("unsupported OTP algorithm '%s'", otpKey.AlgorithmStr)
		}

		otpKey.Algorithm = otpAlgorithmMapping[otpKey.AlgorithmStr]
		otpKey.Digits = otp.Digits(otpKey.DigitsInt)
	}

	// Run GC to remove the plaintext otpKey.Secret from memory
	runtime.GC()

	return otpKeys, nil
}

// GenerateCode generates an OTP token
func (k *OTPKey) GenerateCode() (string, error) {
	if _, ok := otpTypeMapping[k.OTPType]; !ok {
		return "", fmt.Errorf("unsupported OTP type '%s'", k.OTPType)
	}

	secretBuf, err := k.secretEnclave.Open()
	if err != nil {
		memguard.SafePanic(err)
	}

	defer secretBuf.Destroy()

	return otpTypeMapping[k.OTPType](secretBuf.String(), k.Period, k.Digits, k.Algorithm)
}
