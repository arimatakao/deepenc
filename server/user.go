package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func EmptyHandler(c echo.Context) error {
	return c.String(http.StatusOK, "empty handler")
}

type User struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func (s *Server) SignUp(c echo.Context) error {

	return c.String(http.StatusOK, "sign up handler")
}

func (s *Server) VerifySignUp(c echo.Context) error {

	return c.String(http.StatusOK, "verify sign up")
}

func (s *Server) SignIn(c echo.Context) error {

	return c.String(http.StatusOK, "sign in handler")
}
