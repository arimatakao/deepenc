package server

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type jwtCustomClaims struct {
	jwt.RegisteredClaims
}

func newJWT(userId, secret string) (string, error) {
	claims := &jwtCustomClaims{
		jwt.RegisteredClaims{
			ID: userId,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func getUserIdFromJWT(c echo.Context) (string, error) {
	user, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return "", errors.New("can't convert value from context to *jwt.Token")
	}
	claims, ok := user.Claims.(*jwtCustomClaims)
	if !ok {
		return "", errors.New("can't convert jwt token to custom claims")
	}
	return claims.ID, nil
}
