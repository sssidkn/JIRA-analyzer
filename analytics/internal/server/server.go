package server

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sssidkn/JIRA-analyzer/internal/service"
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

func New(service service.Service) *Server {
	e := gin.Default()
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
		api.GET("/graph/get/:taskNumber", s.getGraph)
		api.POST("/graph/make/:taskNumber", s.makeGraph)
		api.DELETE("/graph/delete", s.deleteGraph)
		api.GET("/isAnalyzed", s.isAnalyzed)
		api.GET("/compare/:taskNumber", s.compare)
		//TODO group/ services
	}
}

func (s *Server) Run(port int) error {
	s.httpServer.Addr = ":" + strconv.Itoa(port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
