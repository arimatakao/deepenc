package server

import (
	"context"
	"net/http"

	"github.com/arimatakao/deepenc/cmd/config"
	"github.com/arimatakao/deepenc/server/database"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type Server struct {
	e       *echo.Echo
	db      database.Storager
	cachedb database.Cacher
}

func (s *Server) Init() error {
	s.e = echo.New()
	s.e.HideBanner = true

	s.e.Pre(middleware.RemoveTrailingSlash())
	s.e.Use(middleware.Logger())
	s.e.Logger.SetLevel(log.INFO)

	s.e.RouteNotFound("/*", func(c echo.Context) error {
		return c.NoContent(http.StatusNotFound)
	})

	basePath := s.e.Group("/api")

	// Public routes
	basePath.POST("/signup", s.SignUp)                 // Registration
	basePath.GET("/verify/:token", s.VerifySignUp)     // Verification
	basePath.POST("/signin", s.SignIn)                 // Login
	basePath.GET("/messages/public/:id", EmptyHandler) // Get public message by id
	basePath.POST("/messages/:id", EmptyHandler)       // Get private message by id

	// JWT Auth routes
	messagePath := basePath.Group("/messages")
	messagePath.GET("/public", EmptyHandler) // Get list of public messages with text
	messagePath.GET("", EmptyHandler)        // Get list of user id messages
	messagePath.POST("", EmptyHandler)       // Create message
	messagePath.PUT("/:id", EmptyHandler)    // Update message
	messagePath.DELETE("/:id", EmptyHandler) // Delete message by hand if ttl not set

	// Connect to DB
	db, err := database.NewMainDB(config.MongoURL)
	if err != nil {
		return err
	}
	s.db = db

	// Connect to CacheDB
	cachedb, err := database.NewCacheDB(config.RedisURL)
	if err != nil {
		return err
	}
	s.cachedb = cachedb

	return nil
}

func (s *Server) Run() error {
	return s.e.Start(":" + config.Port)
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.e.Shutdown(ctx); err != nil {
		return err
	}
	if err := s.db.Shutdown(ctx); err != nil {
		return err
	}
	if err := s.cachedb.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}
