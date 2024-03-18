package server

import (
	"context"
	"crypto/sha1"
	"fmt"
	"net/http"

	"github.com/arimatakao/deepenc/server/database"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
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

type cachedUser struct {
	Username       string `redis:"username"`
	HashedPassword string `redis:"password"`
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
	} else if err != nil {
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

	hasher := sha1.New()
	hashedUsername := fmt.Sprintf("%x", hasher.Sum([]byte(u.Username)))

	err = s.r.Get(context.Background(), hashedUsername).Err()
	if err != redis.Nil {
		return c.JSON(http.StatusConflict,
			resp("user is already exist, waiting verification"))
	}

	cUser := cachedUser{
		Username:       u.Username,
		HashedPassword: string(hashedPassword),
	}
	err = s.r.HSet(context.Background(), hashedUsername, cUser).Err()
	if err != nil {
		s.e.Logger.Error(err)
		return c.String(http.StatusInternalServerError, "")
	}

	messageText := fmt.Sprintf("confirm your username by route - /api/verify/%s",
		hashedUsername)
	return c.JSON(http.StatusOK, resp(messageText))
}

func (s *Server) VerifySignUp(c echo.Context) error {
	confirmToken := c.Param("token")
	if confirmToken == "" {
		return c.String(http.StatusBadRequest, "")
	}

	result, err := s.r.HGetAll(context.Background(), confirmToken).Result()
	if err != nil {
		s.e.Logger.Error(err)
		return c.String(http.StatusInternalServerError, "")
	}
	username, ok := result["username"]
	if !ok {
		s.e.Logger.Error(err)
		return c.String(http.StatusInternalServerError, "")
	}
	hashedPassword, ok := result["password"]
	if !ok {
		s.e.Logger.Error(err)
		return c.String(http.StatusInternalServerError, "")
	}

	u := &database.User{
		Username: username,
		Password: hashedPassword,
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

	_, err := s.db.GetUser(u.Username)
	if err == mongo.ErrNoDocuments {
		return c.JSON(http.StatusNotFound, "")
	} else if err != nil {
		s.e.Logger.Error(err)
		c.String(http.StatusInternalServerError, "")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		s.e.Logger.Error(err)
		return c.String(http.StatusInternalServerError, "")
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(u.Password))
	if err != nil {
		return c.String(http.StatusBadRequest, "")
	}

	return c.String(http.StatusOK, "")
}
