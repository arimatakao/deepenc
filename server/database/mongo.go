package database

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MainDB struct {
	client      *mongo.Client
	usersCol    *mongo.Collection
	messagesCol *mongo.Collection
}

func NewMainDB(connectionUrl string) (*MainDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	clientdb, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionUrl))
	if err != nil {
		return nil, err
	}
	err = clientdb.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	database := clientdb.Database("deepenc")
	usersCol := database.Collection("Users")
	messagesCol := database.Collection("Messages")

	db := &MainDB{
		client:      clientdb,
		usersCol:    usersCol,
		messagesCol: messagesCol,
	}

	return db, nil
}

func (d *MainDB) Shutdown(ctx context.Context) error {
	d.usersCol = nil
	d.messagesCol = nil
	return d.client.Disconnect(ctx)
}

func (d MainDB) AddUser(u *User) error {
	ctx := context.Background()
	_, err := d.usersCol.InsertOne(ctx, u)
	return err
}

func (d MainDB) GetUser(username string) (UserOut, error) {
	u := UserOut{}

	ctx := context.Background()
	err := d.usersCol.FindOne(ctx, bson.D{{Key: "username", Value: username}}).Decode(&u)
	if err != nil {
		return UserOut{}, err
	}

	return u, nil
}

func (d MainDB) AddMessage(m *Message) (string, error) {
	ctx := context.Background()
	result, err := d.messagesCol.InsertOne(ctx, m)
	if err != nil {
		return "", err
	}
	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", errors.New("cant convert inserted object id")
	}

	return id.Hex(), nil
}
func (d MainDB) GetMessage(id string) (MessageOut, error) {
	msgId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return MessageOut{}, err
	}

	ctx := context.Background()
	result := d.usersCol.FindOne(ctx, bson.D{{Key: "_id", Value: msgId}})
	if result.Err() != nil {
		return MessageOut{}, result.Err()
	}

	msg := MessageOut{}
	err = result.Decode(&msg)
	if err != nil {
		return MessageOut{}, err
	}

	return msg, nil
}
func (d MainDB) GetLastMessages(skip int) (MessagesOut, error) {
	return MessagesOut{}, nil
}
func (d MainDB) GetUserMessages(ownerId string) (MessagesOut, error) {
	return MessagesOut{}, nil
}
func (d MainDB) UpdateMessage(id string, m Message) error {
	return nil
}
func (d MainDB) DeleteMessage(id string) error {
	return nil
}
