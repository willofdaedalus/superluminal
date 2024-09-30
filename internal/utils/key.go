package utils

import (
	"strings"

	"github.com/sethvargo/go-diceware/diceware"
	"golang.org/x/crypto/bcrypt"
)

func CheckPassphrase(hash, passphrase string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(passphrase)) == nil
}

func GeneratePassphrase() (string, error) {
	list, err := diceware.Generate(6)
	if err != nil {
		return "", err
	}

	return strings.Join(list, " "), nil
}

func HashPassphrase(passphrase string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(passphrase), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}
