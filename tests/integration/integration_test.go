package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestGetProjectsList(t *testing.T) {
	url := "http://host.docker.internal:8080/api/v1/connector/projects?limit=30&page=1&search=A"
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to GET %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("unexpected status %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var result struct {
		Projects []struct {
			Name string `json:"name"`
		} `json:"projects"`
		PageInfo struct {
			PageCount     json.Number `json:"pageCount"`
			ProjectsCount json.Number `json:"projectsCount"`
			CurrentPage   json.Number `json:"currentPage"`
		} `json:"pageInfo"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal json: %v", err)
	}

	projectsCount, _ := result.PageInfo.ProjectsCount.Int64()

	// Проверяем, что вернулось не более 30 проектов
	if len(result.Projects) > 30 {
		t.Fatalf("expected less then 30 projects, got %d", len(result.Projects))
	}

	// Проверяем, что все имена содержат букву "A" (регистр не важен)
	for _, p := range result.Projects {
		if !strings.Contains(strings.ToUpper(p.Name), "A") {
			t.Fatalf("project name does not contain 'A': %s", p.Name)
		}
	}

	if projectsCount < 30 {
		t.Fatalf("expected projectsCount >= 30, got %d", projectsCount)
	}
}

func TestDownloadProject(t *testing.T) {
	projectKey := "AAR"
	projectName := "aardvark"

	body := map[string]string{"project_key": projectKey}
	bodyBytes, _ := json.Marshal(body)
	resp, err := http.Post("http://host.docker.internal:8080/api/v1/connector/updateProject", "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("failed to POST updateProject: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("expected 200 OK, got %d, body: %s", resp.StatusCode, string(data))
	}

	found := false
	for i := 0; i < 20; i++ { // ждём до 10 секунд
		time.Sleep(500 * time.Millisecond)

		getResp, err := http.Get("http://host.docker.internal:8080/api/v1/projects")
		if err != nil {
			continue
		}
		defer getResp.Body.Close()
		if getResp.StatusCode != http.StatusOK {
			continue
		}

		data, err := ioutil.ReadAll(getResp.Body)
		if err != nil {
			continue
		}

		var result struct {
			Data []struct {
				ID   int    `json:"id"`
				Key  string `json:"key"`
				Name string `json:"name"`
			} `json:"data"`
		}

		if err := json.Unmarshal(data, &result); err != nil {
			t.Logf("failed to unmarshal JSON: %v", err)
			continue
		}

		for _, p := range result.Data {
			if strings.EqualFold(p.Key, projectKey) && strings.EqualFold(p.Name, projectName) {
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		t.Fatalf("project %s (%s) not found in database after updateProject", projectKey, projectName)
	}
}
