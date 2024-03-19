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
)

type cfg struct {
	Port       int    `yaml:"port"`
	MongoDBURL string `yaml:"mongodb_url"`
	RedisURL   string `yaml:"redis_url"`
	JWTSecret  string `yaml:"jwt_secret"`
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
		return errors.New("not allowed port value in config")
	}

	if c.JWTSecret == "" {
		return errors.New("jwt_secret field from config is empty")
	}

	Port = strconv.Itoa(c.Port)
	MongoURL = c.MongoDBURL
	RedisURL = c.RedisURL
	JWTSecret = c.JWTSecret

	return nil
}
