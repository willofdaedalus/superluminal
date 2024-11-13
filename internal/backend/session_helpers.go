package backend

import (
	"willofdaedalus/superluminal/internal/utils"
)

// genPassAndHash generates, hashes and returns a new pass and hash
func genPassAndHash(count int) (string, string, error) {
	pass, err := utils.GeneratePassphrase(count)
	if err != nil {
		return "", "", err
	}

	hash, err := utils.HashPassphrase(pass)
	if err != nil {
		return "", "", err
	}

	return pass, hash, nil
}
