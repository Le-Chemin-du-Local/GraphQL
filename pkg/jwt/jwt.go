package jwt

import (
	"strings"
	"time"

	"chemin-du-local.bzh/graphql/internal/config"
	"github.com/golang-jwt/jwt"
)

var (
	SecretKey = []byte(config.Cfg.Settings.AuthSecret)
)

func GenerateToken(id string, role string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = id
	claims["role"] = role
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tokenString, err := token.SignedString(SecretKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseToken(tokenString string) (string, string, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return SecretKey, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id := claims["id"].(string)
		role := claims["role"].(string)

		return id, role, nil
	} else {
		return "", "", err
	}
}
