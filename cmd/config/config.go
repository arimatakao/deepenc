package config

import (
	"errors"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

var Port string

type cfg struct {
	Port int `yaml:"port"`
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

	Port = strconv.Itoa(c.Port)

	return nil
}
