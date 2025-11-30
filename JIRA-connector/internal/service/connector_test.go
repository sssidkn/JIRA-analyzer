package connector

import (
	"context"
	"errors"
	"github.com/sssidkn/jira-connector/internal/models"
	"github.com/sssidkn/jira-connector/pkg/logger"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository мок для Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) SaveProject(ctx context.Context, project models.JiraProject) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *MockRepository) GetProjectInfo(ctx context.Context, projectKey string) (*models.ProjectInfo, error) {
	args := m.Called(ctx, projectKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProjectInfo), args.Error(1)
}

// MockAPIClient мок для APIClient
type MockAPIClient struct {
	mock.Mock
}

func (m *MockAPIClient) UpdateProject(ctx context.Context, projectKey string, lastUpdate time.Time) (*[]models.JiraIssue, error) {
	args := m.Called(ctx, projectKey, lastUpdate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]models.JiraIssue), args.Error(1)
}

func (m *MockAPIClient) GetProject(ctx context.Context, projectKey string) (*models.JiraProject, error) {
	args := m.Called(ctx, projectKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.JiraProject), args.Error(1)
}

func (m *MockAPIClient) GetProjects(ctx context.Context, limit, page int, search string) ([]models.ProjectInfo, error) {
	args := m.Called(ctx, limit, page, search)
	return args.Get(0).([]models.ProjectInfo), args.Error(1)
}

func (m *MockAPIClient) GetBaseURL() string {
	args := m.Called()
	return args.String(0)
}

func createTestProjectInfo() *models.ProjectInfo {
	return &models.ProjectInfo{
		ID:         "10000",
		Key:        "TEST",
		Name:       "Test Project",
		LastUpdate: time.Now().Add(-24 * time.Hour),
		Self:       "https://jira.test.com/rest/api/2/project/TEST",
	}
}

func createTestJiraProject() *models.JiraProject {
	return &models.JiraProject{
		ID:         "10000",
		Key:        "TEST",
		Name:       "Test Project",
		Self:       "https://jira.test.com/rest/api/2/project/TEST",
		LastUpdate: time.Now(),
		Issues: []models.JiraIssue{
			{
				ID:  "10001",
				Key: "TEST-1",
				Fields: models.Fields{
					Summary: "Test Issue",
					Creator: models.JiraUser{
						DisplayName: "Test User",
					},
				},
			},
		},
	}
}

func createTestIssues() *[]models.JiraIssue {
	issues := []models.JiraIssue{
		{
			ID:  "10001",
			Key: "TEST-1",
			Fields: models.Fields{
				Summary: "New Test Issue",
				Creator: models.JiraUser{
					DisplayName: "Test User",
				},
			},
		},
	}
	return &issues
}

func TestNewJiraConnector(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := &MockRepository{}
		mockAPIClient := &MockAPIClient{}
		mockLogger := &logger.TestLogger{}

		connector, err := NewJiraConnector(
			WithRepository(mockRepo),
			WithAPIClient(mockAPIClient),
			WithLogger(mockLogger),
		)

		require.NoError(t, err)
		assert.NotNil(t, connector)
		assert.Equal(t, mockRepo, connector.repo)
		assert.Equal(t, mockAPIClient, connector.apiClient)
		assert.Equal(t, mockLogger, connector.logger)
	})

	t.Run("WithNilRepository", func(t *testing.T) {
		mockAPIClient := &MockAPIClient{}
		mockLogger := &logger.TestLogger{}

		connector, err := NewJiraConnector(
			WithRepository(nil),
			WithAPIClient(mockAPIClient),
			WithLogger(mockLogger),
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repo is nil")
		assert.Nil(t, connector)
	})

	t.Run("WithNilAPIClient", func(t *testing.T) {
		mockRepo := &MockRepository{}
		mockLogger := &logger.TestLogger{}

		connector, err := NewJiraConnector(
			WithRepository(mockRepo),
			WithAPIClient(nil),
			WithLogger(mockLogger),
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "apiClient is nil")
		assert.Nil(t, connector)
	})
}

func TestJiraConnector_GetProjects(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := &MockRepository{}
		mockAPIClient := &MockAPIClient{}
		mockLogger := &logger.TestLogger{}

		connector, err := NewJiraConnector(
			WithRepository(mockRepo),
			WithAPIClient(mockAPIClient),
			WithLogger(mockLogger),
		)
		require.NoError(t, err)

		// Подготовка тестовых данных
		testProjects := []models.ProjectInfo{
			{
				ID:   "10000",
				Key:  "TEST",
				Name: "Test Project",
				Self: "https://jira.test.com/rest/api/2/project/TEST",
			},
			{
				ID:   "10001",
				Key:  "PROJ",
				Name: "Another Project",
				Self: "https://jira.test.com/rest/api/2/project/PROJ",
			},
		}

		// Настройка моков
		mockAPIClient.On("GetProjects", mock.Anything, 10, 1, "test").
			Return(testProjects, nil)

		// Вызов метода
		ctx := context.Background()
		response, err := connector.GetProjects(ctx, 10, 1, "test")

		// Проверки
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Projects, 1) // Только один проект содержит "test" в имени
		assert.Equal(t, "10000", response.Projects[0].Id)
		assert.Equal(t, "TEST", response.Projects[0].Key)
		assert.Equal(t, "Test Project", response.Projects[0].Name)
		assert.Equal(t, int64(1), response.PageInfo.PageCount)
		assert.Equal(t, int64(1), response.PageInfo.ProjectsCount)

		mockAPIClient.AssertExpectations(t)
	})

	t.Run("APIClientError", func(t *testing.T) {
		mockRepo := &MockRepository{}
		mockAPIClient := &MockAPIClient{}
		mockLogger := &logger.TestLogger{}

		connector, err := NewJiraConnector(
			WithRepository(mockRepo),
			WithAPIClient(mockAPIClient),
			WithLogger(mockLogger),
		)
		require.NoError(t, err)

		expectedError := errors.New("API error")

		// Настройка моков
		mockAPIClient.On("GetProjects", mock.Anything, 10, 1, "test").
			Return([]models.ProjectInfo{}, expectedError)

		// Вызов метода
		ctx := context.Background()
		response, err := connector.GetProjects(ctx, 10, 1, "test")

		// Проверки
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, response)

		mockAPIClient.AssertExpectations(t)
	})

	t.Run("ZeroLimit", func(t *testing.T) {
		mockRepo := &MockRepository{}
		mockAPIClient := &MockAPIClient{}
		mockLogger := &logger.TestLogger{}

		connector, err := NewJiraConnector(
			WithRepository(mockRepo),
			WithAPIClient(mockAPIClient),
			WithLogger(mockLogger),
		)
		require.NoError(t, err)

		testProjects := []models.ProjectInfo{
			{
				ID:   "10000",
				Key:  "TEST",
				Name: "Test Project",
				Self: "https://jira.test.com/rest/api/2/project/TEST",
			},
		}

		// Настройка моков
		mockAPIClient.On("GetProjects", mock.Anything, 0, 1, "").
			Return(testProjects, nil)

		// Вызов метода
		ctx := context.Background()
		response, err := connector.GetProjects(ctx, 0, 1, "")

		// Проверки
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "limit cannot be zero")
		assert.Nil(t, response)

		mockAPIClient.AssertExpectations(t)
	})
}

func TestJiraConnector_UpdateProject(t *testing.T) {
	t.Run("ExistingProjectWithNewIssues", func(t *testing.T) {
		mockRepo := &MockRepository{}
		mockAPIClient := &MockAPIClient{}
		mockLogger := &logger.TestLogger{}

		connector, err := NewJiraConnector(
			WithRepository(mockRepo),
			WithAPIClient(mockAPIClient),
			WithLogger(mockLogger),
		)
		require.NoError(t, err)

		projectKey := "TEST"
		projectInfo := createTestProjectInfo()
		testIssues := createTestIssues()
		baseURL := "https://jira.test.com"

		// Настройка моков
		mockRepo.On("GetProjectInfo", mock.Anything, projectKey).
			Return(projectInfo, nil) // Проект найден в БД

		mockAPIClient.On("UpdateProject", mock.Anything, projectKey, projectInfo.LastUpdate).
			Return(testIssues, nil)

		mockAPIClient.On("GetBaseURL").
			Return(baseURL)

		mockRepo.On("SaveProject", mock.Anything, mock.AnythingOfType("models.JiraProject")).
			Return(nil)

		// Вызов метода
		ctx := context.Background()
		result, err := connector.UpdateProject(ctx, projectKey)

		// Проверки
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, projectKey, result.Key)
		assert.Len(t, result.Issues, 1)

		mockRepo.AssertExpectations(t)
		mockAPIClient.AssertExpectations(t)
	})

	t.Run("ExistingProjectNoNewIssues", func(t *testing.T) {
		mockRepo := &MockRepository{}
		mockAPIClient := &MockAPIClient{}
		mockLogger := &logger.TestLogger{}

		connector, err := NewJiraConnector(
			WithRepository(mockRepo),
			WithAPIClient(mockAPIClient),
			WithLogger(mockLogger),
		)
		require.NoError(t, err)

		projectKey := "TEST"
		projectInfo := createTestProjectInfo()
		emptyIssues := &[]models.JiraIssue{}

		// Настройка моков
		mockRepo.On("GetProjectInfo", mock.Anything, projectKey).
			Return(projectInfo, nil)

		mockAPIClient.On("UpdateProject", mock.Anything, projectKey, projectInfo.LastUpdate).
			Return(emptyIssues, nil)

		mockAPIClient.On("GetBaseURL").
			Return("https://jira.test.com")

		// Вызов метода
		ctx := context.Background()
		result, err := connector.UpdateProject(ctx, projectKey)

		// Проверки
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, projectKey, result.Key)
		assert.Len(t, result.Issues, 0) // Нет новых issues

		// SaveProject не должен вызываться когда нет новых issues
		mockRepo.AssertNotCalled(t, "SaveProject")

		mockRepo.AssertExpectations(t)
		mockAPIClient.AssertExpectations(t)
	})

	t.Run("GetProjectInfoError", func(t *testing.T) {
		mockRepo := &MockRepository{}
		mockAPIClient := &MockAPIClient{}
		mockLogger := &logger.TestLogger{}

		connector, err := NewJiraConnector(
			WithRepository(mockRepo),
			WithAPIClient(mockAPIClient),
			WithLogger(mockLogger),
		)
		require.NoError(t, err)

		projectKey := "TEST"
		expectedError := errors.New("database error")

		// Настройка моков
		mockRepo.On("GetProjectInfo", mock.Anything, projectKey).
			Return(nil, expectedError)

		// Вызов метода
		ctx := context.Background()
		result, err := connector.UpdateProject(ctx, projectKey)

		// Проверки
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
		mockAPIClient.AssertNotCalled(t, "GetProject")
		mockAPIClient.AssertNotCalled(t, "UpdateProject")
	})

	t.Run("GetProjectError", func(t *testing.T) {
		mockRepo := &MockRepository{}
		mockAPIClient := &MockAPIClient{}
		mockLogger := &logger.TestLogger{}

		connector, err := NewJiraConnector(
			WithRepository(mockRepo),
			WithAPIClient(mockAPIClient),
			WithLogger(mockLogger),
		)
		require.NoError(t, err)

		projectKey := "TEST"
		expectedError := errors.New("API error")

		// Настройка моков
		mockRepo.On("GetProjectInfo", mock.Anything, projectKey).
			Return(nil, nil)

		mockAPIClient.On("GetProject", mock.Anything, projectKey).
			Return(nil, expectedError)

		// Вызов метода
		ctx := context.Background()
		result, err := connector.UpdateProject(ctx, projectKey)

		// Проверки
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
		mockAPIClient.AssertExpectations(t)
	})

	t.Run("UpdateProjectError", func(t *testing.T) {
		mockRepo := &MockRepository{}
		mockAPIClient := &MockAPIClient{}
		mockLogger := &logger.TestLogger{}

		connector, err := NewJiraConnector(
			WithRepository(mockRepo),
			WithAPIClient(mockAPIClient),
			WithLogger(mockLogger),
		)
		require.NoError(t, err)

		projectKey := "TEST"
		projectInfo := createTestProjectInfo()
		expectedError := errors.New("update error")

		// Настройка моков
		mockRepo.On("GetProjectInfo", mock.Anything, projectKey).
			Return(projectInfo, nil)

		mockAPIClient.On("UpdateProject", mock.Anything, projectKey, projectInfo.LastUpdate).
			Return(nil, expectedError)

		// Вызов метода
		ctx := context.Background()
		result, err := connector.UpdateProject(ctx, projectKey)

		// Проверки
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
		mockAPIClient.AssertExpectations(t)
	})
}
