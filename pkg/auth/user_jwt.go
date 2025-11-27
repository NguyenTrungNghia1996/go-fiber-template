package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

// userBaseSecret returns JWT_SECRET base (shared across tenants).
func userBaseSecret() (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET is not set")
	}
	return secret, nil
}

// tenantSecret derives the per-tenant secret using base + subdomain.
func tenantSecret(subdomain string) (string, error) {
	base, err := userBaseSecret()
	if err != nil {
		return "", err
	}
	return base + subdomain, nil
}

// GenerateUserToken signs a JWT for a unit user using per-tenant key = JWT_SECRET + subdomain.
// Claims include: sub (userID), unit_id, subdomain, is_admin, exp, iat.
func GenerateUserToken(userID string, unitID string, subdomain string, isAdmin bool, ttl time.Duration) (string, time.Time, error) {
	secret, err := tenantSecret(subdomain)
	if err != nil {
		return "", time.Time{}, err
	}
	now := time.Now().UTC()
	exp := now.Add(ttl)
	claims := jwt.MapClaims{
		"sub":       userID,
		"unit_id":   unitID,
		"subdomain": subdomain,
		"is_admin":  isAdmin,
		"exp":       exp.Unix(),
		"iat":       now.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, err
	}
	return s, exp, nil
}

// VerifyUserToken validates a unit user JWT using per-tenant key = JWT_SECRET + subdomain claim.
func VerifyUserToken(tokenString string) (jwt.MapClaims, error) {
	// KeyFunc can read claims to derive the key per subdomain.
	parsed, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		claims, ok := t.Claims.(jwt.MapClaims)
		if !ok {
			return nil, errors.New("invalid claims")
		}
		sdVal, ok := claims["subdomain"]
		if !ok {
			return nil, errors.New("missing subdomain")
		}
		sd, ok := sdVal.(string)
		if !ok || sd == "" {
			return nil, errors.New("invalid subdomain")
		}
		secret, err := tenantSecret(sd)
		if err != nil {
			return nil, err
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
