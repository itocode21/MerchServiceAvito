package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

func SetJWTSecret(secret []byte) {
	jwtSecret = secret
}

func GenerateJWT(username string) (string, error) {
	if len(jwtSecret) == 0 {
		return "", fmt.Errorf("JWT secret not set")
	}
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("Failed generate token: %v", err)
	}
	return signedToken, nil
}

func ValidateJWT(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неверный метод подписи")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return "", fmt.Errorf("ошибка валидации токена: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("недействительный токен")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", fmt.Errorf("неверные данные в токене")
	}

	return username, nil
}
