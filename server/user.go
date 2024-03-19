package server

import (
	"fmt"
	"net/http"

	"github.com/arimatakao/deepenc/server/database"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func EmptyHandler(c echo.Context) error {
	return c.String(http.StatusOK, "empty handler")
}

type systemMessage struct {
	Message string `json:"message"`
}

func resp(text string) *systemMessage {
	return &systemMessage{
		Message: text,
	}
}

func (s *Server) SignUp(c echo.Context) error {
	u := new(database.User)

	if err := c.Bind(u); err != nil {
		return c.String(http.StatusBadRequest, "")
	}

	if u.Username == "" || u.Password == "" {
		return c.String(http.StatusBadRequest, "")
	}

	_, err := s.db.GetUser(u.Username)
	if err != mongo.ErrNoDocuments {
		return c.JSON(http.StatusConflict, resp("user is already exist"))
	} else if err != nil && err != mongo.ErrNoDocuments {
		s.e.Logger.Error(err)
		return c.String(http.StatusInternalServerError, "")
	}

	if len(u.Password) < 8 && len(u.Password) > 33 {
		return c.JSON(http.StatusBadRequest,
			resp("password should contain more than 7 symbols"))
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		s.e.Logger.Error(err)
		return c.String(http.StatusInternalServerError, "")
	}

	token, err := s.cachedb.AddUser(u.Username, string(hashedPassword))
	if err != nil {
		s.e.Logger.Error(err)
		return c.String(http.StatusInternalServerError, "")
	}

	messageText := fmt.Sprintf("confirm your username by route - /api/verify/%s", token)
	return c.JSON(http.StatusOK, resp(messageText))
}

func (s *Server) VerifySignUp(c echo.Context) error {
	confirmToken := c.Param("token")
	if confirmToken == "" {
		return c.String(http.StatusBadRequest, "")
	}

	u, err := s.cachedb.GetUser(confirmToken)
	if err != nil {
		s.e.Logger.Error(err)
		return c.String(http.StatusInternalServerError, "")
	}

	if err = s.db.AddUser(u); err != nil {
		s.e.Logger.Error(err)
		c.String(http.StatusInternalServerError, "")
	}

	return c.String(http.StatusCreated, "")
}

func (s *Server) SignIn(c echo.Context) error {
	u := new(database.User)

	if err := c.Bind(u); err != nil {
		return c.String(http.StatusBadRequest, "")
	}

	if u.Username == "" || u.Password == "" {
		return c.String(http.StatusBadRequest, "")
	}

	userDocument, err := s.db.GetUser(u.Username)
	if err == mongo.ErrNoDocuments {
		return c.JSON(http.StatusNotFound, "")
	} else if err != nil {
		s.e.Logger.Error(err)
		c.String(http.StatusInternalServerError, "")
	}

	err = bcrypt.CompareHashAndPassword([]byte(userDocument.Password), []byte(u.Password))
	if err != nil {
		return c.String(http.StatusBadRequest, "")
	}

	return c.String(http.StatusOK, "")
}
