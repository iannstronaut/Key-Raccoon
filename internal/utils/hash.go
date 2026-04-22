package utils

import "golang.org/x/crypto/bcrypt"

const (
	CostOfHashing = 8
)

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), CostOfHashing)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

func VerifyPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func CheckPasswordHash(password, hashedPassword string) bool {
	return VerifyPassword(password, hashedPassword) == nil
}
