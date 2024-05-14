package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/arimatakao/deepenc/cmd/config"
	"github.com/arimatakao/deepenc/server"
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

	srv := new(server.Server)
	err = srv.Init()
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		if err := srv.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Error occurred while running server: ", err.Error())
		} else {
			log.Println("Shutdown server")
		}
	}()

	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server shutdown with error: ", err.Error())
	}

	log.Println("Shutdown is successful")
}
