package utils

import (
	"strings"

	"github.com/sethvargo/go-diceware/diceware"
	"golang.org/x/crypto/bcrypt"
)

// normalizePassphrase ensures consistent handling of passphrases
func normalizePassphrase(passphrase string) string {
	return strings.TrimSpace(passphrase)
}

func CheckPassphrase(hash, passphrase string) bool {
	normalizedPassphrase := normalizePassphrase(passphrase)
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(normalizedPassphrase)) == nil
}

func GeneratePassphrase() (string, error) {
	list, err := diceware.Generate(1)
	if err != nil {
		return "", err
	}
	return normalizePassphrase(strings.Join(list, " ")), nil
}

func HashPassphrase(passphrase string) (string, error) {
	normalizedPassphrase := normalizePassphrase(passphrase)
	hash, err := bcrypt.GenerateFromPassword([]byte(normalizedPassphrase), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
