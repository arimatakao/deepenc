package server

import (
	"net/http"

	"github.com/arimatakao/deepenc/server/database"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

type Message struct {
	Content       string `json:"content"`
	IsPrivate     bool   `json:"is_private"`
	EncodingType  string `json:"encoding_type"`
	Password      string `json:"password"`
	OnlyOwnerView bool   `json:"only_owner_view"`
	IsAnon        bool   `json:"is_anon"`
	IsOneTime     bool   `json:"is_one_time"`
}

func (m Message) toDatabaseFormat(userId string) *database.Message {
	return &database.Message{
		OwnerId:       userId,
		Content:       m.Content,
		IsPrivate:     m.IsPrivate,
		EncodingType:  m.EncodingType,
		Password:      m.Password,
		OnlyOwnerView: m.OnlyOwnerView,
		IsAnon:        m.IsAnon,
		IsOneTime:     m.IsOneTime,
	}
}

func (s *Server) CreateMessage(c echo.Context) error {
	userId, err := getUserIdFromJWT(c)
	if err != nil {
		c.Logger().Error(err)
		return c.String(http.StatusInternalServerError, "")
	}

	msg := new(Message)
	if err := c.Bind(msg); err != nil {
		return c.String(http.StatusBadRequest, "")
	}

	if msg.Content == "" {
		return c.String(http.StatusBadRequest, "")
	}

	mFormat := msg.toDatabaseFormat(userId)

	resultId, err := s.db.AddMessage(mFormat)
	if err != nil {
		c.Logger().Error(err)
		return c.String(http.StatusInternalServerError, "")
	}

	c.Logger().Info("added new message: " + resultId)

	return c.String(http.StatusCreated, "")
}

func (s *Server) GetPublicMessage(c echo.Context) error {
	msgId := c.Param("id")
	if msgId == "" {
		return c.String(http.StatusBadRequest, "")
	}

	msg, err := s.db.GetMessage(msgId)
	if err == mongo.ErrNoDocuments {
		return c.String(http.StatusNotFound, "")
	}
	if err != nil {
		c.Logger().Error(err)
		return c.String(http.StatusInternalServerError, "")
	}

	if msg.IsPrivate || msg.Password != "" || msg.EncodingType != "plaintext" {
		return c.String(http.StatusNotFound, "")
	}

	if msg.IsAnon {
		msg.OwnerId = ""
	}

	return c.JSON(http.StatusOK, msg)
}
