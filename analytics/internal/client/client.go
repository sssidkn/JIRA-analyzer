package client

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/sssidkn/JIRA-analyzer/pkg/api/connector"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Client() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewJiraConnectorClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test GetProjects
	fmt.Println("Getting projects...")
	projectsResp, err := client.GetProjects(ctx, &pb.GetProjectsRequest{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		log.Fatalf("GetProjects failed: %v", err)
	}

	fmt.Printf("Found %d projects:\n", len(projectsResp.Projects))
	for _, p := range projectsResp.Projects {
		fmt.Printf("- %s (%s): %s\n", p.Key, p.Name, p.Description)
	}

	// Test UpdateProject
	fmt.Println("\nUpdating project...")
	updateResp, err := client.UpdateProject(ctx, &pb.UpdateProjectRequest{
		ProjectKey: "PROJ1",
	})
	if err != nil {
		log.Fatalf("UpdateProject failed: %v", err)
	}

	fmt.Printf("Update result: %v\n", updateResp.Success)
	fmt.Printf("Updated project: %s (%s)\n", updateResp.Project.Key, updateResp.Project.Name)
}
