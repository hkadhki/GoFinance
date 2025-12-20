package jwt

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secret = []byte(getEnv("JWT_SECRET", "dev-secret"))

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func Generate(userID string) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func Validate(tokenStr string) (string, bool, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(t *jwt.Token) (any, error) {
			return secret, nil
		},
	)

	if err != nil || !token.Valid {
		return "", false, err
	}

	claims := token.Claims.(*Claims)
	return claims.UserID, true, nil
}
