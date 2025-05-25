package server

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sssidkn/JIRA-analyzer/internal/repository"
)

func (s *Server) getProjects(c *gin.Context) {
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		limit = 20
	}
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		offset = 0
	}

	ctx := c.Request.Context()
	response, err := s.service.GetProjects(ctx, limit, offset)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, response)
}

func (s *Server) getProject(c *gin.Context) {
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx := c.Request.Context()
	response, err := s.service.GetProject(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotExist) {
			c.String(http.StatusNotFound, err.Error())
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, response)
}

func (s *Server) deleteProject(c *gin.Context) {
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request.Context()
	err = s.service.DeleteProject(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotExist) {
			c.String(http.StatusNotFound, err.Error())
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}
