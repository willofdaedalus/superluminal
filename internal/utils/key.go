package utils

import (
	"strings"

	"github.com/sethvargo/go-diceware/diceware"
	"golang.org/x/crypto/bcrypt"
)

// checks the client entered passphrase with the hashed passphrase
func CheckPassphrase(hash, passphrase string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(passphrase)) == nil
}

// generates a new diceware passphrase
func GeneratePassphrase() (string, error) {
	list, err := diceware.Generate(6)
	if err != nil {
		return "", err
	}

	return strings.Join(list, " "), nil
}

// returns a hashed passphrase
func HashPassphrase(passphrase string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(passphrase), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}
