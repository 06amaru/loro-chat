package utils

import (
	"server/models"
	"time"

	"github.com/golang-jwt/jwt"
)

type JwtClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func MakeToken(username string) (*models.Credential, error) {
	jwtClaims := JwtClaims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * 24 * 60 * 60).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)

	ss, err := token.SignedString(MySigningKey)
	return &models.Credential{Username: username, Token: ss}, err
}
