package utils

import (
	"time"

	"github.com/golang-jwt/jwt"
)

type JwtClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func MakeToken(username string) (string, error) {
	jwtClaims := JwtClaims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * 24 * 60 * 60).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)

	ss, err := token.SignedString(MySigningKey)
	return ss, err
}
