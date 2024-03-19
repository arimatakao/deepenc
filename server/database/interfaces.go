package database

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type cachedUser struct {
	Username       string `redis:"username"`
	HashedPassword string `redis:"password"`
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserOut struct {
	Id       primitive.ObjectID `bson:"_id"`
	Username string             `json:"username"`
	Password string             `json:"password"`
}

type UsersDB interface {
	AddUser(u *User) error
	GetUser(username string) (UserOut, error)
}

type Cacher interface {
	AddUser(username string, hashedPassword string) (token string, err error)
	GetUser(token string) (*User, error)
	Shutdown(context.Context) error
}

type Message struct {
	OwnerId       string `json:"owner_id"`
	Content       string `json:"content"`
	IsPrivate     bool   `json:"is_private"`
	EncodingType  string `json:"encoding_type"`
	Password      string `json:"password"`
	OnlyOwnerView bool   `json:"only_owner_view"`
	IsAnon        bool   `json:"is_anon"`
	IsOneTime     bool   `json:"is_one_time"`
}

type MessageOut struct {
	Id            primitive.ObjectID `bson:"_id"`
	OwnerId       string             `json:"owner_id"`
	Content       string             `json:"content"`
	IsPrivate     bool               `json:"is_private"`
	EncodingType  string             `json:"encoding_type"`
	Password      string             `json:"password"`
	OnlyOwnerView bool               `json:"only_owner_view"`
	IsAnon        bool               `json:"is_anon"`
	IsOneTime     bool               `json:"is_one_time"`
}

type MessagesOut []MessageOut

type MessagesDB interface {
	AddMessage(m *Message) (id string, err error)
	GetMessage(id string) (MessageOut, error)
	GetLastMessages(skip int) (MessagesOut, error)
	GetUserMessages(ownerId string) (MessagesOut, error)
	UpdateMessage(id string, m Message) error
	DeleteMessage(id string) error
}

type Storager interface {
	UsersDB
	MessagesDB
	Shutdown(context.Context) error
}
