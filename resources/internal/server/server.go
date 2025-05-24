package server

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sssidkn/JIRA-analyzer/internal/service"
	"github.com/sssidkn/JIRA-analyzer/pkg/logger"
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

func New(service service.Service, l *logger.Logger) *Server {
	e := gin.New()
	e.Use(gin.Recovery())
	e.Use(logger.Middleware(l))
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
		//TODO issues, history
	}
}

func (s *Server) Run(port int) error {
	s.httpServer.Addr = ":" + strconv.Itoa(port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
