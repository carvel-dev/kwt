package registry

// Based on https://github.com/cloudfoundry/config-server/blob/5b33cbdec5929168ce338ec827e866ec333ed889/types/password_generator.go

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

const DefaultPasswordLength = 20

type PasswordGenerator struct{}

type PasswordParams struct {
	Length int `yaml:"length"`
}

func NewPasswordGenerator() PasswordGenerator {
	return PasswordGenerator{}
}

func (PasswordGenerator) Generate(params PasswordParams) (string, error) {
	if params.Length < 0 {
		return "", fmt.Errorf("Length param cannot be negative")
	}
	if params.Length == 0 {
		params.Length = DefaultPasswordLength
	}

	lengthLetterRunes := big.NewInt(int64(len(letterRunes)))
	passwordRunes := make([]rune, params.Length)

	for i := range passwordRunes {
		index, err := rand.Int(rand.Reader, lengthLetterRunes)
		if err != nil {
			return "", err
		}

		passwordRunes[i] = letterRunes[index.Int64()]
	}

	return string(passwordRunes), nil
}
