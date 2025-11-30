package server

import (
	"context"
	"errors"
	"github.com/sssidkn/jira-connector/internal/models"
	connectorApi "github.com/sssidkn/jira-connector/pkg/api/connector"
	"github.com/sssidkn/jira-connector/pkg/logger"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

// MockService мок для Service интерфейса
type MockService struct {
	mock.Mock
}

func (m *MockService) UpdateProject(ctx context.Context, projectKey string) (*models.JiraProject, error) {
	args := m.Called(ctx, projectKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.JiraProject), args.Error(1)
}

func (m *MockService) GetProjects(ctx context.Context, limit, page int, search string) (*connectorApi.GetProjectsResponse, error) {
	args := m.Called(ctx, limit, page, search)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*connectorApi.GetProjectsResponse), args.Error(1)
}

// bufConnListener создает in-memory соединение для тестов
const bufSize = 1024 * 1024

func createTestServer(t *testing.T, service Service) (*GRPCServer, *grpc.ClientConn, func()) {
	lis := bufconn.Listen(bufSize)

	testLogger := &logger.TestLogger{}

	server := NewGRPCServer(
		WithService(service),
		WithLogger(testLogger),
	)

	grpcServer := grpc.NewServer()
	connectorApi.RegisterJiraConnectorServer(grpcServer, server)

	go func() {
		if err := grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			t.Errorf("gRPC server failed: %v", err)
		}
	}()

	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithInsecure(),
	)
	require.NoError(t, err)

	cleanup := func() {
		conn.Close()
		grpcServer.Stop()
		lis.Close()
	}

	return server, conn, cleanup
}

func TestNewGRPCServer(t *testing.T) {
	t.Run("WithNilService", func(t *testing.T) {
		testLogger := &logger.TestLogger{}

		// Это не должно паниковать, но сервер будет без сервиса
		server := NewGRPCServer(
			WithService(nil),
			WithLogger(testLogger),
		)

		assert.NotNil(t, server)
		assert.Nil(t, server.service)
	})
}

func TestGRPCServer_UpdateProject(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService := &MockService{}
		_, conn, cleanup := createTestServer(t, mockService)
		defer cleanup()

		client := connectorApi.NewJiraConnectorClient(conn)

		expectedProject := &models.JiraProject{
			ID:   "10000",
			Key:  "TEST",
			Name: "Test Project",
			Self: "https://jira.test.com/projects/TEST",
		}

		// Настройка моков
		mockService.On("UpdateProject", mock.Anything, "TEST").
			Return(expectedProject, nil)

		// Вызов метода
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		response, err := client.UpdateProject(ctx, &connectorApi.UpdateProjectRequest{
			ProjectKey: "TEST",
		})

		// Проверки
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.True(t, response.Success)
		assert.Equal(t, "10000", response.Project.Id)
		assert.Equal(t, "TEST", response.Project.Key)
		assert.Equal(t, "Test Project", response.Project.Name)
		assert.Equal(t, "https://jira.test.com/projects/TEST", response.Project.Url)

		mockService.AssertExpectations(t)
	})

	t.Run("ServiceError", func(t *testing.T) {
		mockService := &MockService{}
		_, conn, cleanup := createTestServer(t, mockService)
		defer cleanup()

		client := connectorApi.NewJiraConnectorClient(conn)

		expectedError := errors.New("service error")

		// Настройка моков
		mockService.On("UpdateProject", mock.Anything, "TEST").
			Return(nil, expectedError)

		// Вызов метода
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		response, err := client.UpdateProject(ctx, &connectorApi.UpdateProjectRequest{
			ProjectKey: "TEST",
		})

		// Проверки
		assert.Error(t, err)
		assert.Nil(t, response)

		// Проверяем что ошибка правильно передается через gRPC
		grpcStatus, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unknown, grpcStatus.Code())
		assert.Contains(t, grpcStatus.Message(), "service error")

		mockService.AssertExpectations(t)
	})
}

func TestGRPCServer_GetProjects(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService := &MockService{}
		_, conn, cleanup := createTestServer(t, mockService)
		defer cleanup()

		client := connectorApi.NewJiraConnectorClient(conn)

		expectedResponse := &connectorApi.GetProjectsResponse{
			Projects: []*connectorApi.JiraProject{
				{
					Id:   "10000",
					Key:  "TEST",
					Name: "Test Project",
					Url:  "https://jira.test.com/projects/TEST",
				},
			},
			PageInfo: &connectorApi.PageInfo{
				PageCount:     1,
				ProjectsCount: 1,
			},
		}

		// Настройка моков
		mockService.On("GetProjects", mock.Anything, 10, 1, "test").
			Return(expectedResponse, nil)

		// Вызов метода
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		response, err := client.GetProjects(ctx, &connectorApi.GetProjectsRequest{
			Limit:  10,
			Page:   1,
			Search: "test",
		})

		// Проверки
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Projects, 1)
		assert.Equal(t, int64(1), response.PageInfo.PageCount)
		assert.Equal(t, int64(1), response.PageInfo.ProjectsCount)
		assert.Equal(t, "TEST", response.Projects[0].Key)

		mockService.AssertExpectations(t)
	})

	t.Run("ServiceError", func(t *testing.T) {
		mockService := &MockService{}
		_, conn, cleanup := createTestServer(t, mockService)
		defer cleanup()

		client := connectorApi.NewJiraConnectorClient(conn)

		expectedError := errors.New("get projects error")

		// Настройка моков
		mockService.On("GetProjects", mock.Anything, 10, 1, "test").
			Return(nil, expectedError)

		// Вызов метода
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		response, err := client.GetProjects(ctx, &connectorApi.GetProjectsRequest{
			Limit:  10,
			Page:   1,
			Search: "test",
		})

		// Проверки
		assert.Error(t, err)
		assert.Nil(t, response)

		grpcStatus, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unknown, grpcStatus.Code())
		assert.Contains(t, grpcStatus.Message(), "get projects error")

		mockService.AssertExpectations(t)
	})
}

func TestGRPCServer_StartAndStop(t *testing.T) {
	t.Run("StartSuccess", func(t *testing.T) {
		mockService := &MockService{}
		testLogger := &logger.TestLogger{}

		server := NewGRPCServer(
			WithService(mockService),
			WithLogger(testLogger),
		)

		// Запускаем сервер на случайном порту
		listener, err := net.Listen("tcp", "localhost:0")
		require.NoError(t, err)

		server.server = grpc.NewServer()
		connectorApi.RegisterJiraConnectorServer(server.server, server)

		// Тестируем остановку сервера
		done := make(chan bool)
		go func() {
			server.server.Serve(listener)
			done <- true
		}()

		// Даем серверу время запуститься
		time.Sleep(100 * time.Millisecond)

		// Останавливаем сервер
		server.server.Stop()

		// Ждем завершения
		select {
		case <-done:
			// Сервер успешно остановился
		case <-time.After(2 * time.Second):
			t.Fatal("Server stop timeout")
		}
	})

	t.Run("StopWithoutStart", func(t *testing.T) {
		mockService := &MockService{}
		testLogger := &logger.TestLogger{}

		server := NewGRPCServer(
			WithService(mockService),
			WithLogger(testLogger),
		)

		// Остановка без запуска не должна паниковать
		assert.NotPanics(t, func() {
			server.Stop()
		})
	})
}

func TestGRPCServer_ConcurrentRequests(t *testing.T) {
	t.Run("MultipleRequests", func(t *testing.T) {
		mockService := &MockService{}
		_, conn, cleanup := createTestServer(t, mockService)
		defer cleanup()

		client := connectorApi.NewJiraConnectorClient(conn)

		project := &models.JiraProject{
			ID:   "10000",
			Key:  "TEST",
			Name: "Test Project",
			Self: "https://jira.test.com/projects/TEST",
		}

		projectsResponse := &connectorApi.GetProjectsResponse{
			Projects: []*connectorApi.JiraProject{
				{
					Id:   "10000",
					Key:  "TEST",
					Name: "Test Project",
					Url:  "https://jira.test.com/projects/TEST",
				},
			},
			PageInfo: &connectorApi.PageInfo{
				PageCount:     1,
				ProjectsCount: 1,
			},
		}

		// Настройка моков для нескольких вызовов
		mockService.On("UpdateProject", mock.Anything, "TEST").Return(project, nil).Times(2)
		mockService.On("GetProjects", mock.Anything, 10, 1, "").Return(projectsResponse, nil).Times(2)

		ctx := context.Background()
		errors := make(chan error, 4)

		// Запускаем несколько горутин
		for i := 0; i < 2; i++ {
			go func() {
				_, err := client.UpdateProject(ctx, &connectorApi.UpdateProjectRequest{
					ProjectKey: "TEST",
				})
				errors <- err
			}()
		}

		for i := 0; i < 2; i++ {
			go func() {
				_, err := client.GetProjects(ctx, &connectorApi.GetProjectsRequest{
					Limit:  10,
					Page:   1,
					Search: "",
				})
				errors <- err
			}()
		}

		// Ждем завершения всех горутин
		for i := 0; i < 4; i++ {
			err := <-errors
			assert.NoError(t, err)
		}

		mockService.AssertExpectations(t)
	})
}
