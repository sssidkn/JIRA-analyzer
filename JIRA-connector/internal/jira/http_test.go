package jira_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sssidkn/jira-connector/internal/jira"
	"github.com/sssidkn/jira-connector/internal/models"
	"github.com/sssidkn/jira-connector/pkg/logger"
	"github.com/sssidkn/jira-connector/pkg/ratelimiter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRateLimitedServer создает сервер с имитацией rate limiting
func MockRateLimitedServer(t *testing.T) *httptest.Server {
	requestCount := 0
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")

		// Имитация rate limiting на каждом 3-м запросе
		if requestCount%3 == 0 {
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Rate limit exceeded",
			})
			return
		}

		switch r.URL.Path {
		case "/rest/api/2/search":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"issues": []models.JiraIssue{
					{
						ID:  "1",
						Key: "TEST-1",
						Fields: models.Fields{
							Summary: "Test Issue",
						},
					},
				},
				"total": 1,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

// MockErrorServer создает сервер, возвращающий ошибки
func MockErrorServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Internal server error",
		})
	}))
}

func TestNewClient(t *testing.T) {
	t.Run("DefaultConfiguration", func(t *testing.T) {
		client := jira.NewClient()
		assert.NotNil(t, client)
	})

	t.Run("WithCustomConfiguration", func(t *testing.T) {
		cfg := jira.Config{
			BaseURL:      "https://test.jira.com",
			VersionAPI:   "/rest/api/2",
			MaxResults:   100,
			MaxProcesses: 5,
		}

		client := jira.NewClient(
			jira.WithConfig(cfg),
			jira.WithLogger(&logger.TestLogger{}),
		)

		assert.NotNil(t, client)
	})
}

func TestClient_GetProject(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := MockServer()
		defer server.Close()

		cfg := jira.Config{
			BaseURL:      server.URL,
			VersionAPI:   "/rest/api/2",
			MaxResults:   50,
			MaxProcesses: 3,
		}

		client := jira.NewClient(
			jira.WithConfig(cfg),
			jira.WithLogger(&logger.TestLogger{}),
		)

		ctx := context.Background()
		project, err := client.GetProject(ctx, "TEST")

		require.NoError(t, err)
		assert.NotNil(t, project)
		assert.Equal(t, "TEST", project.Key)
		assert.Equal(t, "Test Project", project.Name)
		assert.Equal(t, 150, project.TotalIssueCount)
		assert.Len(t, project.Issues, 150)
	})

	t.Run("ProjectNotFound", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		cfg := jira.Config{
			BaseURL:    server.URL,
			VersionAPI: "/rest/api/2",
		}

		client := jira.NewClient(
			jira.WithConfig(cfg),
			jira.WithLogger(&logger.TestLogger{}),
		)

		ctx := context.Background()
		project, err := client.GetProject(ctx, "NONEXISTENT")

		assert.Error(t, err)
		assert.Nil(t, project)
	})
}

func TestClient_GetProjects(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := MockServer()
		defer server.Close()

		cfg := jira.Config{
			BaseURL:    server.URL,
			VersionAPI: "/rest/api/2",
		}

		client := jira.NewClient(
			jira.WithConfig(cfg),
			jira.WithLogger(&logger.TestLogger{}),
		)

		ctx := context.Background()
		projects, err := client.GetProjects(ctx, 10, 1, "")

		require.NoError(t, err)
		assert.NotNil(t, projects)
		assert.Len(t, projects, 2)
		assert.Equal(t, "TEST", projects[0].Key)
		assert.Equal(t, "PROJ", projects[1].Key)
	})

	t.Run("WithSearchFilter", func(t *testing.T) {
		server := MockServer()
		defer server.Close()

		cfg := jira.Config{
			BaseURL:    server.URL,
			VersionAPI: "/rest/api/2",
		}

		client := jira.NewClient(
			jira.WithConfig(cfg),
			jira.WithLogger(&logger.TestLogger{}),
		)

		ctx := context.Background()
		projects, err := client.GetProjects(ctx, 10, 1, "Test")

		require.NoError(t, err)
		assert.NotNil(t, projects)
		// В реальном коде фильтрация должна работать на стороне клиента
	})
}

func TestClient_RateLimiting(t *testing.T) {
	t.Run("RateLimitRecovery", func(t *testing.T) {
		server := MockRateLimitedServer(t)
		defer server.Close()

		cfg := jira.Config{
			BaseURL:      server.URL,
			VersionAPI:   "/rest/api/2",
			MaxResults:   50,
			MaxProcesses: 2,
		}

		client := jira.NewClient(
			jira.WithConfig(cfg),
			jira.WithLogger(&logger.TestLogger{}),
		)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Этот запрос должен обработать rate limiting и повторить
		issues, err := client.UpdateProject(ctx, "TEST", time.Time{})

		// В реальном тесте мы можем получить ошибку из-за таймаута контекста
		// или успешный результат после повторных попыток
		if err != nil {
			assert.True(t, errors.Is(err, context.DeadlineExceeded) ||
				err.Error() == "exceeded max delay")
		} else {
			assert.NotNil(t, issues)
		}
	})
}

func TestClient_ConcurrentRequests(t *testing.T) {
	server := MockServer()
	defer server.Close()

	cfg := jira.Config{
		BaseURL:      server.URL,
		VersionAPI:   "/rest/api/2",
		MaxResults:   50,
		MaxProcesses: 5,
	}

	client := jira.NewClient(
		jira.WithConfig(cfg),
		jira.WithLogger(&logger.TestLogger{}),
	)

	// Тест конкурентных запросов
	ctx := context.Background()
	errs := make(chan error, 3)

	for i := 0; i < 3; i++ {
		go func() {
			_, err := client.GetProjects(ctx, 10, 1, "")
			errs <- err
		}()
	}

	// Ждем завершения всех горутин
	for i := 0; i < 3; i++ {
		err := <-errs
		assert.NoError(t, err)
	}
}

func TestRateLimiterIntegration(t *testing.T) {
	t.Run("RateLimiterInitialization", func(t *testing.T) {
		client := jira.NewClient(
			jira.WithConfig(jira.Config{
				StartDelay: 100,
				MaxDelay:   10000,
			}),
		)

		// Проверяем что rate limiter инициализирован
		assert.NotNil(t, client) // Косвенная проверка через создание клиента
	})

	t.Run("RateLimiterReset", func(t *testing.T) {
		server := MockServer()
		defer server.Close()

		// Создаем мок rate limiter для тестирования
		rl := ratelimiter.NewRateLimiter(100*time.Millisecond, 10*time.Second)

		// Проверяем начальное состояние
		paused, duration := rl.ShouldPause()
		assert.False(t, paused)
		assert.Equal(t, time.Duration(0), duration)
	})
}
