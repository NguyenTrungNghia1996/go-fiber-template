package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

func adminSecret() (string, error) {
	secret := os.Getenv("JWT_SECRET_ADMIN")
	if secret == "" {
		return "", errors.New("JWT_SECRET_ADMIN is not set")
	}
	return secret, nil
}

// GenerateToken signs a JWT for the given subject with an expiration TTL using JWT_SECRET_ADMIN.
func GenerateToken(subject string, ttl time.Duration) (string, time.Time, error) {
	secret, err := adminSecret()
	if err != nil {
		return "", time.Time{}, err
	}
	now := time.Now().UTC()
	exp := now.Add(ttl)
	claims := jwt.MapClaims{
		"sub": subject,
		"exp": exp.Unix(),
		"iat": now.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, err
	}
	return s, exp, nil
}

// VerifyToken validates a JWT string using JWT_SECRET_ADMIN and returns its claims if valid.
func VerifyToken(tokenString string) (jwt.MapClaims, error) {
	secret, err := adminSecret()
	if err != nil {
		return nil, err
	}
	parsed, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}
