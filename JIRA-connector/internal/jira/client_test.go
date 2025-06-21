package jira_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"jira-connector/internal/jira"
	"jira-connector/internal/models"
	"jira-connector/pkg/logger"
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
		RetryCount:     3,
		MaxResults:     100,
	}

	client := jira.NewClient(cfg)
	client.SetLogger(logger.NewTestLogger())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.UpdateProject(context.Background(), "TEST")
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
		RetryCount:     3,
		MaxResults:     100,
	}

	client := jira.NewClient(cfg)
	client.SetLogger(logger.NewTestLogger())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.UpdateProject(context.Background(), "TEST")
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
		RetryCount:     3,
		MaxResults:     100,
	}

	client := jira.NewClient(cfg)
	client.SetLogger(logger.NewTestLogger())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.UpdateProject(context.Background(), "TEST")
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

	client := jira.NewClient(cfg)
	client.SetLogger(logger.NewTestLogger())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetProjects(context.Background(), 10, 1, "test")
		if err != nil {
			b.Fatalf("GetProjects failed: %v", err)
		}
	}
}
