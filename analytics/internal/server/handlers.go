package server

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sssidkn/JIRA-analyzer/internal/repository"
)

func (s *Server) getGraph(c *gin.Context) {
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
	issues, err := s.service.GetTask(c.Request.Context(), task, key)
	if err != nil {
		if errors.Is(err, repository.ErrNotExistProject) {
			c.String(http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, repository.ErrNotExistData) {
			c.String(http.StatusNotFound, err.Error())
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, issues)
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
	ctx := c.Request.Context()
	issues, err := s.service.MakeTask(ctx, task, key)
	if err != nil {
		if errors.Is(err, repository.ErrNotExistProject) {
			c.String(http.StatusNotFound, err.Error())
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, issues)
}

func (s *Server) deleteGraph(c *gin.Context) {
	key := c.Query("project")
	if key == "" {
		c.String(http.StatusBadRequest, "no key")
		return
	}

	ok, err := s.service.DeleteTasks(c.Request.Context(), key)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, ok)
}

func (s *Server) isAnalyzed(c *gin.Context) {
	key := c.Query("project")
	if key == "" {
		c.String(http.StatusBadRequest, "no key")
		return
	}

	ok, err := s.service.IsAnalyzed(c.Request.Context(), key)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, ok)
}

func (s *Server) compare(c *gin.Context) {

}
