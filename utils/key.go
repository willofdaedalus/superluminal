package utils

import (
	"strings"

	"github.com/sethvargo/go-diceware/diceware"
	"golang.org/x/crypto/bcrypt"
)

// NormalizePassphrase ensures consistent handling of passphrases
func NormalizePassphrase(passphrase string) string {
	return strings.TrimSpace(passphrase)
}

func CheckPassphrase(hash, passphrase string) bool {
	normalizedPassphrase := NormalizePassphrase(passphrase)
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(normalizedPassphrase)) == nil
}

func GeneratePassphrase() (string, error) {
	list, err := diceware.Generate(1)
	if err != nil {
		return "", err
	}
	return NormalizePassphrase(strings.Join(list, " ")), nil
}

func HashPassphrase(passphrase string) (string, error) {
	normalizedPassphrase := NormalizePassphrase(passphrase)
	hash, err := bcrypt.GenerateFromPassword([]byte(normalizedPassphrase), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
