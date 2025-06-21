package server

import (
	"context"
	connector "jira-connector/internal/service"
	connectorApi "jira-connector/pkg/api/connector"
)

type GRPCServer struct {
	connectorApi.UnimplementedJiraConnectorServer
	service *connector.JiraConnector
}

func NewGRPCServer(service *connector.JiraConnector) *GRPCServer {
	return &GRPCServer{service: service}
}

func (s *GRPCServer) UpdateProject(ctx context.Context,
	req *connectorApi.UpdateProjectRequest) (*connectorApi.UpdateProjectResponse, error) {
	project, err := s.service.UpdateProject(ctx, req.GetProjectKey())
	if err != nil {
		return nil, err
	}

	return &connectorApi.UpdateProjectResponse{
		Project: &connectorApi.JiraProject{
			Id:   project.ID,
			Key:  project.Key,
			Name: project.Name,
		},
		Success: true,
	}, nil
}

func (s *GRPCServer) GetProjects(ctx context.Context, req *connectorApi.GetProjectsRequest) (*connectorApi.GetProjectsResponse, error) {
	projects, err := s.service.GetProjects(ctx, int(req.GetPageSize()), int(req.GetPage()), req.GetSearch())
	if err != nil {
		return nil, err
	}

	response := &connectorApi.GetProjectsResponse{
		Projects: make([]*connectorApi.JiraProject, 0, len(projects)),
		PageInfo: &connectorApi.PageInfo{},
	}

	for _, p := range projects {
		response.Projects = append(response.Projects, &connectorApi.JiraProject{
			Id:   p.ID,
			Key:  p.Key,
			Name: p.Name,
		})
	}

	return response, nil
}
