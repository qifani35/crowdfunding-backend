package auth

import (
	"errors"
	"os"

	"github.com/dgrijalva/jwt-go"
)

type Service interface {
	GenerateToken(userID int) (string, error)
	ValidateToken(tokenString string) (*jwt.Token, error)
}

type jwtService struct {
}

func getSecretKey() []byte {
	key := os.Getenv("SECRET_KEY")
	if key == "" {
		// Fallback for development only — override in production via env
		key = "CHANGE_ME_IN_PRODUCTION"
	}
	return []byte(key)
}

func NewService() Service {
	return &jwtService{}
}

func (s *jwtService) GenerateToken(userID int) (string, error) {
	claims := jwt.MapClaims{}
	claims["user_id"] = userID

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(getSecretKey())

	if err != nil {
		return signedToken, err
	}

	return signedToken, nil
}

func (s *jwtService) ValidateToken(encodedToken string) (*jwt.Token, error) {
	token, err := jwt.Parse(encodedToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid token")
		}

		return getSecretKey(), nil
	})

	if err != nil {
		return token, err
	}

	return token, nil

}
