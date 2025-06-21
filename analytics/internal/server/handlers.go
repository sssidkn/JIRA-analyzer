package server

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetGraph godoc
// @Summary Get analytical data for a specific task
// @Description Retrieves graph data for the specified task number and project key
// @Produce json
// @Param taskNumber path int true "Task number to retrieve graph for"
// @Param project query string true "Project key identifier"
// @Success 200 {object} dto.IssueTaskOne "Данные для задачи типа 1"
// @Success 200 {object} dto.IssueTaskTwo "Данные для задачи типа 2"
// @Failure 400 {string} string "Invalid task number or missing project key"
// @Failure 404 {string} string "Task or project not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/graph/get/{taskNumber} [get]
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
		if errors.Is(err, errNotExist) {
			c.String(http.StatusNotFound, err.Error())
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, issues)
}

// MakeGraph godoc
// @Summary Generate analytical data for a task
// @Description Creates and returns analytical graph data for the specified task
// @Produce json
// @Param taskNumber path int true "Task number to generate graph for"
// @Param project query string true "Project key identifier"
// @Success 200 {object} dto.IssueTaskOne "Данные для задачи типа 1"
// @Success 200 {object} dto.IssueTaskTwo "Данные для задачи типа 2"
// @Failure 400 {string} string "Invalid task number or missing project key"
// @Failure 404 {string} string "Task or project not found"
// @Failure 500 {string} string "Internal server error
// @Router /api/v1/graph/make/{taskNumber} [post]
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
		if errors.Is(err, errNotExist) {
			c.String(http.StatusNotFound, err.Error())
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, issues)
}

// DeleteGraph godoc
// @Summary Delete all graph data for a project
// @Description Removes all analytical graph data associated with the specified project
// @Produce json
// @Param project query string true "Project key identifier"
// @Success 200 {boolean} bool "True if deletion was successful"
// @Failure 400 {string} string "Missing project key"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/graph/delete [delete]
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

// IsAnalyzed godoc
// @Summary Check if project has been analyzed
// @Description Verifies whether analytical data exists for the specified project
// @Produce json
// @Param project query string true "Project key identifier"
// @Success 200 {boolean} bool "True if project has been analyzed"
// @Failure 400 {string} string "Missing project key"
// @Failure 404 {string} string "Project not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/isAnalyzed [get]
func (s *Server) isAnalyzed(c *gin.Context) {
	key := c.Query("project")
	if key == "" {
		c.String(http.StatusBadRequest, "no key")
		return
	}

	ok, err := s.service.IsAnalyzed(c.Request.Context(), key)
	if err != nil {
		if errors.Is(err, errNotExist) {
			c.String(http.StatusNotFound, err.Error())
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, ok)
}

// Compare godoc
// @Summary Compare analytical data for a task
// @Description Retrieves comparison data for the specified task across projects
// @Produce json
// @Param taskNumber path int true "Task number to compare"
// @Param project query string true "Comma-separated project keys to compare"
// @Success 200 {object} dto.ComparisonTaskOne "Данные для задачи типа 1"
// @Success 200 {object} dto.ComparisonTaskTwo "Данные для задачи типа 2"
// @Failure 400 {string} string "Invalid task number or missing project keys"
// @Failure 404 {string} string "Task or projects not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/compare/{taskNumber} [get]
func (s *Server) compare(c *gin.Context) {
	task, err := strconv.Atoi(c.Params.ByName("taskNumber"))
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	keys := c.Query("project")
	if keys == "" {
		c.String(http.StatusBadRequest, "no key")
		return
	}

	comparisons, err := s.service.Compare(c.Request.Context(), task, keys)
	if err != nil {
		if errors.Is(err, errNotExist) {
			c.String(http.StatusNotFound, err.Error())
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, comparisons)
}

var errNotExist = errors.New("does not exist")
