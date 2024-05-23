package server

import (
	"errors"
	"net/http"

	"github.com/arimatakao/deepenc/cmd/config"
	"github.com/arimatakao/deepenc/server/database"
	"github.com/arimatakao/deepenc/utils"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

const (
	MIN_CONTENT_SIZE  = 16
	MIN_PASSWORD_SIZE = 8

	MAX_CONTENT_SIZE = 2000
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

func (m Message) isValid() bool {
	if len(m.Content) > MAX_CONTENT_SIZE ||
		m.Content == "" {
		return false
	}

	switch m.EncodingType {
	case "plaintext":
		if m.Password != "" {
			return false
		}
	case "password":
		if len(m.Password) < MIN_PASSWORD_SIZE {
			return false
		}
	case "internal":
		if len(m.Content) < MIN_CONTENT_SIZE {
			return false
		}
	case "aes":
		if len(m.Content) < MIN_CONTENT_SIZE {
			return false
		}
		if len(m.Password) < MIN_PASSWORD_SIZE {
			return false
		}
	default:
		return false
	}

	return true
}

func (m *Message) formatToEncodingType() error {
	switch m.EncodingType {
	case "plaintext":
		m.Password = ""
	case "password":
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(m.Password),
			bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		m.Password = string(hashedPassword)
		m.IsPrivate = true
	case "internal":
		encrypted, err := utils.EncryptAES256(config.AESInternalKey, m.Content)
		if err != nil {
			return err
		}
		m.Content = encrypted
		m.Password = ""
		m.IsPrivate = true
	case "aes":
		encrypted, err := utils.EncryptAES256([]byte(m.Password), m.Content)
		if err != nil {
			return err
		}
		m.Content = encrypted
		m.Password = ""
		m.IsPrivate = true
	default:
		return errors.New("unknown encoding_type")
	}
	return nil
}

func toOutputFormat(dbmsg database.MessageOut) *Message {
	return &Message{
		Content:       dbmsg.Content,
		IsPrivate:     dbmsg.IsPrivate,
		EncodingType:  dbmsg.EncodingType,
		Password:      dbmsg.Password,
		OnlyOwnerView: dbmsg.OnlyOwnerView,
		IsAnon:        dbmsg.IsAnon,
		IsOneTime:     dbmsg.IsOneTime,
	}
}

type InputPassword struct {
	Password string `json:"password"`
}

func (s *Server) CreateMessage(c echo.Context) error {
	userId, err := getUserIdFromJWT(c)
	if err != nil {
		return c.String(http.StatusBadRequest, "")
	}

	msg := new(Message)
	if err := c.Bind(msg); err != nil {
		return c.String(http.StatusBadRequest, "")
	}

	if !msg.isValid() {
		return c.String(http.StatusBadRequest, "")
	}

	if err := msg.formatToEncodingType(); err != nil {
		c.Logger().Error(err)
		return c.String(http.StatusInternalServerError, "")
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

	if msg.IsPrivate ||
		msg.Password != "" ||
		msg.EncodingType != "plaintext" ||
		msg.OnlyOwnerView {
		return c.String(http.StatusNotFound, "")
	}

	if msg.IsAnon {
		msg.OwnerId = ""
	}

	if msg.IsOneTime {
		if err = s.db.DeleteMessage(msgId); err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusNotFound, "")
		}
	}

	return c.JSON(http.StatusOK, msg)
}

func (s *Server) GetUserMessagesList(c echo.Context) error {
	userId, err := getUserIdFromJWT(c)
	if err != nil {
		return c.String(http.StatusBadRequest, "")
	}

	messages, err := s.db.GetUserMessages(userId)
	if err == mongo.ErrNoDocuments {
		return c.String(http.StatusNotFound, "")
	}
	if err != nil {
		c.Logger().Error(err)
		return c.String(http.StatusInternalServerError, "")
	}

	return c.JSON(http.StatusOK, messages)
}

func (s *Server) GetPublicMessagesList(c echo.Context) error {
	messages, err := s.db.GetLastPublicMessages(10)
	if err == mongo.ErrNoDocuments {
		return c.String(http.StatusNotFound, "")
	}
	if err != nil {
		c.Logger().Error(err)
		return c.String(http.StatusInternalServerError, "")
	}
	for _, v := range messages {
		if v.IsAnon {
			v.OwnerId = ""
		}
	}

	return c.JSON(http.StatusOK, messages)
}

func (s *Server) UpdateMessage(c echo.Context) error {
	userId, err := getUserIdFromJWT(c)
	if err != nil {
		return c.String(http.StatusBadRequest, "")
	}

	msgId := c.Param("id")
	if msgId == "" {
		return c.String(http.StatusBadRequest, "")
	}

	msg := new(Message)
	if err := c.Bind(msg); err != nil {
		return c.String(http.StatusBadRequest, "")
	}

	if !msg.isValid() {
		return c.String(http.StatusBadRequest, "")
	}

	_, err = s.db.GetMessage(msgId)
	if err == mongo.ErrNoDocuments {
		return c.String(http.StatusNotFound, "")
	}
	if err != nil {
		c.Logger().Error(err)
		return c.String(http.StatusInternalServerError, "")
	}

	if err = msg.formatToEncodingType(); err != nil {
		c.Logger().Error(err)
		return c.String(http.StatusInternalServerError, "")
	}

	mFormat := msg.toDatabaseFormat(userId)

	err = s.db.UpdateMessage(msgId, mFormat)
	if err != nil {
		c.Logger().Error(err)
		return c.String(http.StatusBadRequest, "")
	}

	return c.String(http.StatusNoContent, "")
}

func (s *Server) DeleteMessage(c echo.Context) error {
	userId, err := getUserIdFromJWT(c)
	if err != nil {
		return c.String(http.StatusBadRequest, "")
	}

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

	if msg.OwnerId != userId {
		return c.String(http.StatusBadRequest, "")
	}

	err = s.db.DeleteMessage(msgId)
	if err != nil {
		c.Logger().Error(err)
		return c.String(http.StatusInternalServerError, "")
	}

	return c.String(http.StatusNoContent, "")
}

func (s *Server) GetPrivateMessage(c echo.Context) error {
	msgId := c.Param("id")
	if msgId == "" {
		return c.String(http.StatusBadRequest, "")
	}

	input := new(InputPassword)
	if err := c.Bind(input); err != nil {
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

	if !msg.IsPrivate || msg.OnlyOwnerView {
		return c.String(http.StatusNotFound, "")
	}

	switch msg.EncodingType {
	case "password":
		err = bcrypt.CompareHashAndPassword([]byte(msg.Password), []byte(input.Password))
		if err != nil {
			return c.String(http.StatusNotFound, "")
		}
	case "internal":
		decrypted, err := utils.DecryptAES256(config.AESInternalKey, msg.Content)
		if err != nil {
			c.Logger().Error(err)
			return c.String(http.StatusInternalServerError, "")
		}
		msg.Content = decrypted
	case "aes":
		decrypted, err := utils.DecryptAES256([]byte(input.Password), msg.Content)
		if err != nil {
			c.Logger().Warn(err)
			return c.String(http.StatusNotFound, "")
		}
		msg.Content = decrypted
		msg.Password = ""
	}

	msgResp := toOutputFormat(msg)

	if msgResp.IsAnon {
		msg.OwnerId = ""
	}

	if msgResp.IsOneTime {
		if err = s.db.DeleteMessage(msgId); err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusNotFound, "")
		}
	}

	return c.JSON(http.StatusOK, msgResp)
}
