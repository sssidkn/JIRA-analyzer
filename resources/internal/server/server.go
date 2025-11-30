package server

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sssidkn/resources/internal/service"
	"github.com/sssidkn/resources/pkg/logger"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	engine     *gin.Engine
	service    service.Service
	httpServer *http.Server
}

// @title Jira-Analyzer API
// @version 1.0
// @description API for analyze Jira projects
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8084
// @BasePath /api
// @schemes http

func New(service service.Service, l *logger.Logger, timeout time.Duration) *Server {
	e := gin.New()
	e.Use(gin.Recovery())
	e.Use(timeoutMiddleware(timeout))
	e.Use(logger.Middleware(l))
	e.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
	s := &Server{
		engine:  e,
		service: service,
		httpServer: &http.Server{
			Handler: e,
		},
	}
	s.registerRouters()
	return s
}

func (s *Server) registerRouters() {
	s.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	api := s.engine.Group("/api/v1")
	{
		api.GET("/projects", s.getProjects)
		api.GET("projects/:id", s.getProject)
		api.DELETE("/projects/:id", s.deleteProject)
		api.GET("/issues/:id", s.getIssue)
		api.GET("/issues/by-project/:projectId", s.getIssuesByProject)
		api.GET("/histories/by-issue/:issueId", s.getHistoryByIssue)
		api.GET("/histories/by-author/:authorId", s.getHistoryByAuthor)
	}
}

func (s *Server) Run(port int) error {
	s.httpServer.Addr = ":" + strconv.Itoa(port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func timeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		done := make(chan struct{})

		go func() {
			c.Next()
			close(done)
		}()

		select {
		case <-done:
			return
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{
					"error": "Request timeout",
				})
			}
		}
	}
}
