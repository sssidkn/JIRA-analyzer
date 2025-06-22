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

// getProjects godoc
// @Summary Получить список проектов
// @Description Возвращает список проектов с пагинацией
// @Tags Projects
// @Produce json
// @Param limit query int false "Лимит записей (по умолчанию 20)"
// @Param offset query int false "Смещение (по умолчанию 0)"
// @Success 200 {object} models.PaginatedResponse
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /api/v1/projects [get]
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

// getProject godoc
// @Summary Получить проект по ID
// @Description Возвращает проект по указанному идентификатору
// @Tags Projects
// @Produce json
// @Param id path int true "ID проекта"
// @Success 200 {object} models.Response
// @Failure 400 {string} string "Неверный ID проекта"
// @Failure 404 {string} string "Проект не найден"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /api/v1/projects/{id} [get]
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

// deleteProject godoc
// @Summary Удалить проект
// @Description Удаляет проект по указанному идентификатору
// @Tags Projects
// @Param id path int true "ID проекта"
// @Success 204 "Проект успешно удален"
// @Failure 400 {string} string "Неверный ID проекта"
// @Failure 404 {string} string "Проект не найден"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /api/v1/projects/{id} [delete]
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

// getIssue godoc
// @Summary Получить задачу по ID
// @Description Возвращает задачу по указанному идентификатору
// @Tags Issues
// @Produce json
// @Param id path int true "ID задачи"
// @Success 200 {object} models.Response
// @Failure 400 {string} string "Неверный ID задачи"
// @Failure 404 {string} string "Задача не найдена"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /api/v1/issues/{id} [get]
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

// getIssuesByProject godoc
// @Summary Получить задачи проекта
// @Description Возвращает список задач для указанного проекта с пагинацией
// @Tags Issues
// @Produce json
// @Param projectId path int true "ID проекта"
// @Param limit query int false "Лимит записей (по умолчанию 20)"
// @Param offset query int false "Смещение (по умолчанию 0)"
// @Success 200 {object} models.PaginatedResponse
// @Failure 400 {string} string "Неверные параметры запроса"
// @Failure 404 {string} string "Проект не найден"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /api/v1/issues/by-project/{projectId} [get]
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

// getHistoryByIssue godoc
// @Summary Получить историю изменений задачи
// @Description Возвращает историю изменений для указанной задачи
// @Tags History
// @Produce json
// @Param issueId path int true "ID задачи"
// @Success 200 {object} models.Response
// @Failure 400 {string} string "Неверный ID задачи"
// @Failure 404 {string} string "Задача не найдена"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /api/v1/histories/by-issue/{issueId} [get]
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

// getHistoryByAuthor godoc
// @Summary Получить историю изменений автора
// @Description Возвращает историю изменений, сделанных указанным автором
// @Tags History
// @Produce json
// @Param authorId path int true "ID автора"
// @Success 200 {object} models.Response
// @Failure 400 {string} string "Неверный ID автора"
// @Failure 404 {string} string "Автор не найден"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /api/v1/histories/by-author/{authorId} [get]
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
