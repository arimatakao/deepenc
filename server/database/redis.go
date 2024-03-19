package database

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type CacheDB struct {
	r *redis.Client
}

func NewCacheDB(connectionUrl string) (*CacheDB, error) {
	opt, err := redis.ParseURL(connectionUrl)
	if err != nil {
		return nil, err
	}

	r := redis.NewClient(opt)

	err = r.Ping(context.Background()).Err()
	if err != nil {
		return nil, err
	}

	return &CacheDB{
		r: r,
	}, nil
}

func (c CacheDB) Shutdown(context.Context) error {
	return c.r.Close()
}

func (c CacheDB) AddUser(username string, hashedPassword string) (string, error) {
	hasher := sha1.New()
	hashedUsername := fmt.Sprintf("%x", hasher.Sum([]byte(username)))

	isExist, err := c.r.Exists(context.Background(), hashedUsername).Result()
	if err != nil {
		return "", err
	}

	if isExist == 1 {
		return "", errors.New("user is already exist in cache")
	}

	cUser := cachedUser{
		Username:       username,
		HashedPassword: hashedPassword,
	}
	err = c.r.HSet(context.Background(), hashedUsername, cUser).Err()
	if err != nil {
		return "", err
	}

	return hashedUsername, nil
}

func (c CacheDB) GetUser(token string) (*User, error) {
	result, err := c.r.HGetAll(context.Background(), token).Result()
	if err != nil {
		return nil, err
	}
	username, ok := result["username"]
	if !ok {
		return nil, errors.New("username field in hset not exist")
	}
	hashedPassword, ok := result["password"]
	if !ok {
		return nil, errors.New("password field in hset not exist")
	}

	err = c.r.Del(context.Background(), token).Err()
	if err != nil {
		return nil, err
	}

	return &User{
		Username: username,
		Password: hashedPassword,
	}, nil
}
