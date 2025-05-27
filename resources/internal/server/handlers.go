package server

import (
	"context"
	"errors"
	"fmt"
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
	ctx = context.WithValue(ctx, "url", getFullURL(c))
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
	ctx = context.WithValue(ctx, "url", getFullURL(c))
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

func (s *Server) getIssue(c *gin.Context) {
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request.Context()
	ctx = context.WithValue(ctx, "url", getFullURL(c))
	response, err := s.service.GetIssue(ctx, id)
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

func (s *Server) getIssuesByProject(c *gin.Context) {
	projectId, err := strconv.Atoi(c.Params.ByName("projectId"))
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
	}
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		limit = 20
	}
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		offset = 0
	}

	ctx := c.Request.Context()
	ctx = context.WithValue(ctx, "url", getFullURL(c))
	response, err := s.service.GetIssuesByProject(ctx, projectId, limit, offset)
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

func (s *Server) getHistoryByIssue(c *gin.Context) {
	issueId, err := strconv.Atoi(c.Params.ByName("issueId"))
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request.Context()
	ctx = context.WithValue(ctx, "url", getFullURL(c))
	response, err := s.service.GetHistoryByIssue(ctx, issueId)
	if err != nil {
		if errors.Is(err, repository.ErrNotExist) {
			c.String(http.StatusNotFound, err.Error())
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
	}
	c.JSON(http.StatusOK, response)
}

func (s *Server) getHistoryByAuthor(c *gin.Context) {
	authorId, err := strconv.Atoi(c.Params.ByName("authorId"))
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request.Context()
	ctx = context.WithValue(ctx, "url", getFullURL(c))
	response, err := s.service.GetHistoryByAuthor(ctx, authorId)
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

func getFullURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	} else if c.Request.TLS != nil {
		scheme = "https"
	}

	host := c.Request.Host
	if forwardedHost := c.Request.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
		host = forwardedHost
	}

	return fmt.Sprintf("%s://%s%s", scheme, host, c.Request.URL.RequestURI())
}
