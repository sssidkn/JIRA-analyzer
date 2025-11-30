package jira_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sssidkn/jira-connector/internal/jira"
	"github.com/sssidkn/jira-connector/internal/models"
	"github.com/sssidkn/jira-connector/pkg/logger"
)

func MockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/project/TEST":
			w.Write([]byte(`{"id": "10000", "key": "TEST", "name": "Test Project"}`))
		case "/search":
			time.Sleep(50 * time.Millisecond)

			issues := make([]models.JiraIssue, 0)
			for i := 0; i < 20000; i++ {
				issues = append(issues, models.JiraIssue{
					ID:     fmt.Sprintf("%d", i),
					Key:    "TEST-1",
					Fields: models.Fields{Summary: "Test Issue"},
				})
			}

			response := struct {
				Issues []models.JiraIssue `json:"issues"`
				Total  int                `json:"total"`
			}{
				Issues: issues,
				Total:  20000,
			}

			data, _ := json.Marshal(response)
			w.Write(data)
		case "/project":
			projects := []models.ProjectInfo{
				{ID: "10000", Key: "TEST", Name: "Test Project"},
			}
			data, _ := json.Marshal(projects)
			w.Write(data)
		case "/rest/api/2/project/TEST":
			project := models.JiraProject{
				ID:   "10000",
				Key:  "TEST",
				Name: "Test Project",
			}
			json.NewEncoder(w).Encode(project)

		case "/rest/api/2/search":
			// Обработка параметров запроса
			jql := r.URL.Query().Get("jql")
			maxResults := r.URL.Query().Get("maxResults")
			startAt := r.URL.Query().Get("startAt")

			// Для теста общего количества issues
			if maxResults == "0" {
				if jql == "project=TEST" {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"total": 150,
					})
				} else if jql == "project=TEST AND updated > \"2023/01/01\"" {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"total": 25,
					})
				}
				return
			}

			// Для теста получения issues с пагинацией
			var issues []models.JiraIssue
			pageSize := 50
			if maxResults != "" {
				fmt.Sscanf(maxResults, "%d", &pageSize)
			}

			start := 0
			if startAt != "" {
				fmt.Sscanf(startAt, "%d", &start)
			}

			for i := start; i < start+pageSize && i < 150; i++ {
				issues = append(issues, models.JiraIssue{
					ID:  fmt.Sprintf("%d", i),
					Key: fmt.Sprintf("TEST-%d", i),
					Fields: models.Fields{
						Summary: fmt.Sprintf("Test Issue %d", i),
						Creator: models.JiraUser{
							DisplayName: "Test User",
						},
					},
				})
			}

			response := map[string]interface{}{
				"issues": issues,
				"total":  150,
			}
			json.NewEncoder(w).Encode(response)

		case "/rest/api/2/project":
			projects := []models.ProjectInfo{
				{ID: "10000", Key: "TEST", Name: "Test Project"},
				{ID: "10001", Key: "PROJ", Name: "Another Project"},
			}
			json.NewEncoder(w).Encode(projects)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func BenchmarkGetProject_OneConnectionPerGoroutine(b *testing.B) {
	server := MockServer()
	defer server.Close()

	cfg := jira.Config{
		BaseURL:        server.URL,
		MaxConnections: 100,
		MaxProcesses:   100,
		MaxResults:     100,
	}

	client := jira.NewClient(jira.WithConfig(cfg), jira.WithLogger(logger.NewTestLogger()))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.UpdateProject(context.Background(), "TEST", time.Time{})
		if err != nil {
			b.Fatalf("UpdateProject failed: %v", err)
		}
	}
}

func BenchmarkGetProject_SingleSharedConnection(b *testing.B) {
	server := MockServer()
	defer server.Close()

	cfg := jira.Config{
		BaseURL:        server.URL,
		MaxConnections: 1,
		MaxProcesses:   100,
		MaxResults:     100,
	}

	client := jira.NewClient(jira.WithConfig(cfg), jira.WithLogger(logger.NewTestLogger()))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.UpdateProject(context.Background(), "TEST", time.Time{})
		if err != nil {
			b.Fatalf("UpdateProject failed: %v", err)
		}
	}
}

func BenchmarkGetProject_PooledConnections(b *testing.B) {
	server := MockServer()
	defer server.Close()

	cfg := jira.Config{
		BaseURL:        server.URL,
		MaxConnections: 3,
		MaxProcesses:   100,
		MaxResults:     100,
	}

	client := jira.NewClient(jira.WithConfig(cfg), jira.WithLogger(logger.NewTestLogger()))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.UpdateProject(context.Background(), "TEST", time.Time{})
		if err != nil {
			b.Fatalf("UpdateProject failed: %v", err)
		}
	}
}

func BenchmarkGetProjects(b *testing.B) {
	server := MockServer()
	defer server.Close()

	cfg := jira.Config{
		BaseURL: server.URL,
	}

	client := jira.NewClient(jira.WithConfig(cfg), jira.WithLogger(logger.NewTestLogger()))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetProjects(context.Background(), 10, 1, "test")
		if err != nil {
			b.Fatalf("GetProjects failed: %v", err)
		}
	}
}
