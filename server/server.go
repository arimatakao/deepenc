package server

import (
	"context"
	"net/http"
	"time"

	"github.com/arimatakao/deepenc/cmd/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	e *echo.Echo
}

func (s *Server) Init() {
	s.e = echo.New()
	s.e.HideBanner = true

	s.e.Pre(middleware.RemoveTrailingSlash())
	s.e.Use(middleware.Logger())

	s.e.RouteNotFound("/*", func(c echo.Context) error {
		return c.NoContent(http.StatusNotFound)
	})

	basePath := s.e.Group("/api")

	// Public routes
	basePath.POST("/signup", s.SignUp)                 // Registration
	basePath.GET("/verify", s.VerifySignUp)            // Verification
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
}

func (s *Server) Run() error {
	return s.e.Start(":" + config.Port)
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	return s.e.Shutdown(ctx)
}
