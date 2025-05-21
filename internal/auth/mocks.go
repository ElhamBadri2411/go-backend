package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type MockJWTAuthenticator struct{}

const secret = "secret"

var testClaims = jwt.MapClaims{
	"sub": int64(21),
	"exp": time.Now().Add(time.Minute * 5).Unix(),
	"iat": time.Now().Unix(),
	"nbf": time.Now().Unix(),
	"iss": "test-aud",
	"aud": "test-aud",
}

func NewMockJWTAuthenticator() *MockJWTAuthenticator {
	return &MockJWTAuthenticator{}
}

func (m *MockJWTAuthenticator) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, testClaims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (m *MockJWTAuthenticator) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
}
