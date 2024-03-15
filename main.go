package main

import (
	"flag"
	"log"

	"github.com/arimatakao/deepenc/cmd/config"
)

var pathToConfig *string = flag.String("config", "./config.yaml", "path to config yaml file")

func init() {
	flag.Parse()
}

func main() {
	err := config.LoadConfig(*pathToConfig)
	if err != nil {
		log.Fatal(err)
	}
}
