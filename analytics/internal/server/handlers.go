package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (s *Server) getGraph(c *gin.Context) {

}

func (s *Server) makeGraph(c *gin.Context) {
	task, err := strconv.Atoi(c.Params.ByName("taskNumber"))
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	key := c.Query("project")
	if key == "" {
		c.String(http.StatusBadRequest, "no key")
		return
	}
	issues, err := s.service.MakeTask(c.Request.Context(), task, key)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, issues)
}

func (s *Server) deleteGraph(c *gin.Context) {

}

func (s *Server) isAnalyzed(c *gin.Context) {

}

func (s *Server) compare(c *gin.Context) {

}
