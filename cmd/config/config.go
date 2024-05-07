package config

import (
	"errors"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

var (
	Port      string
	MongoURL  string
	RedisURL  string
	JWTSecret string
	AESInternalKey []byte
)

type cfg struct {
	Port       int    `yaml:"port"`
	MongoDBURL string `yaml:"mongodb_url"`
	RedisURL   string `yaml:"redis_url"`
	JWTSecret  string `yaml:"jwt_secret"`
	AESInternalKey string `yaml:"aes_internal_key"`
}

func LoadConfig(pathToYaml string) error {
	data, err := os.ReadFile(pathToYaml)
	if err != nil {
		return err
	}

	c := new(cfg)
	if err = yaml.Unmarshal(data, c); err != nil {
		return err
	}

	if c.Port <= 0 {
		return errors.New("port value from config is not allowed")
	}

	if c.JWTSecret == "" {
		return errors.New("jwt_secret field from config is empty")
	}

	if len(c.AESInternalKey) < 8 {
		return errors.New("aes_internal_key field is shorter than 8 symbols in config")
	}

	Port = strconv.Itoa(c.Port)
	MongoURL = c.MongoDBURL
	RedisURL = c.RedisURL
	JWTSecret = c.JWTSecret
	AESInternalKey = []byte(c.AESInternalKey)

	return nil
}
