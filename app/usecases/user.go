package usecases

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/Brigant/GoPetPorject/app/enteties"
	"github.com/golang-jwt/jwt"
)

const (
	signingKey = "sdFWlnxb13t&r"
	tokenTTL   = 1 * time.Hour
)

type User struct {
	Repo Repository
}

func NewUser(repo Repository) User {
	return User{Repo: repo}
}

// The function CreateUser represents bissness logic layer
// and  execute some manipulation with data before
// saving  this data in the repository.
func (u User) CreateUser(user enteties.User) (string, error) {
	user.Password = u.sha256(user.Password)

	id, err := u.Repo.AddUser(user)
	if err != nil {
		return "", fmt.Errorf("error occures while AddUser: %w", err)
	}

	return id, nil
}

// The function sha256 returns the checksum of the string using SHA256 hash algorithms.
func (u User) sha256(someString string) string {
	h := sha256.Sum256([]byte(someString))

	return fmt.Sprintf("%x", h)
}

type tokenClaims struct {
	jwt.StandardClaims
	Id string
}

// The function GenerateToken represents bissness logic layer
// and  generate token.
func (u User) GenerateToken(phone int, password string) (string, error) {
	userId, err := u.Repo.GetUser(phone, u.sha256(password))
	if err != nil {
		return "", fmt.Errorf("error occures while GenerateToken: %w", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		userId,
	})

	tk, err := token.SignedString([]byte(signingKey))
	if err != nil {
		return "", fmt.Errorf("cannot get SignetString token: %w", err)
	}

	return tk, nil
}

func (u User) ParseToken(accesToken string) (string, error) {
	token, err := jwt.ParseWithClaims(accesToken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing metod")
		}

		return []byte(signingKey), nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return "", errors.New("token claims are not of type *tokenClaims")
	}

	return claims.Id, nil
}
