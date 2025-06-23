package repository

import (
	"context"
	"fmt"
	"jira-connector/internal/models"
	"jira-connector/pkg/logger"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Project = models.JiraProject

type ProjectRepository struct {
	db     *pgxpool.Pool
	logger logger.Logger
}

func (p *ProjectRepository) SetLogger(logger logger.Logger) {
	p.logger = logger
}

func NewProjectRepository(db *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{
		db: db,
	}
}

func (p *ProjectRepository) GetProjectInfo(ctx context.Context, projectKey string) (*models.ProjectInfo, error) {
	var exists bool
	err := p.db.QueryRow(ctx,
		`SELECT 
        EXISTS(SELECT 1 FROM Projects WHERE key = $1)`,
		projectKey,
	).Scan(&exists)
	if !exists {
		return nil, nil
	}
	var pi = &models.ProjectInfo{}
	err = p.db.QueryRow(ctx,
		`SELECT id, key, title, lastUpdate FROM Projects WHERE key = $1`,
		projectKey,
	).Scan(&pi.ID, &pi.Key, &pi.Name, &pi.LastUpdate)
	if err != nil {
		return nil, err
	}
	return pi, nil
}

func (p *ProjectRepository) SaveProject(ctx context.Context, project Project) error {
	tx, err := p.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
        INSERT INTO Projects (id, title, key, lastUpdate) 
        VALUES ($1, $2, $3, $4) 
        ON CONFLICT (key) DO UPDATE SET lastUpdate = EXCLUDED.lastUpdate
    `, project.ID, project.Name, project.Key, project.LastUpdate)
	if err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	authorSet := make(map[string]struct{})
	var statusChanges []StatusChangeData

	for _, issue := range project.Issues {
		authorSet[issue.Fields.Creator.DisplayName] = struct{}{}
		if issue.Fields.Assignee.DisplayName != "" {
			authorSet[issue.Fields.Assignee.DisplayName] = struct{}{}
		}

		for _, history := range issue.Changelogs.Histories {
			authorSet[history.Author.DisplayName] = struct{}{}
		}
	}

	authorNames := make([]string, 0, len(authorSet))
	for name := range authorSet {
		authorNames = append(authorNames, name)
	}

	_, err = tx.Exec(ctx, `
        INSERT INTO Author (name) 
        SELECT unnest($1::text[]) 
        ON CONFLICT DO NOTHING
    `, authorNames)
	if err != nil {
		return fmt.Errorf("failed to batch insert authors: %w", err)
	}

	rows, err := tx.Query(ctx, `
        SELECT name, id FROM Author 
        WHERE name = ANY($1)
    `, authorNames)
	if err != nil {
		return fmt.Errorf("failed to get author IDs: %w", err)
	}
	defer rows.Close()

	authorIDs := make(map[string]int)
	for rows.Next() {
		var name string
		var id int
		if err := rows.Scan(&name, &id); err != nil {
			return fmt.Errorf("failed to scan author ID: %w", err)
		}
		authorIDs[name] = id
	}

	issueBatch := &pgx.Batch{}
	issueKeys := make([]string, 0, len(project.Issues))
	issueKeyToID := make(map[string]int)

	for _, issue := range project.Issues {
		issueKeys = append(issueKeys, issue.Key)

		issueBatch.Queue(`
            INSERT INTO Issue (
                projectId, authorId, assigneeId, key, summary, description, 
                type, priority, status, createdTime, closedTime, updatedTime, timeSpent
            ) VALUES (
                $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
            ) ON CONFLICT (key) DO UPDATE SET
                summary = EXCLUDED.summary,
                description = EXCLUDED.description,
                type = EXCLUDED.type,
                priority = EXCLUDED.priority,
                status = EXCLUDED.status,
                updatedTime = EXCLUDED.updatedTime,
                closedTime = EXCLUDED.closedTime,
                timeSpent = EXCLUDED.timeSpent
            RETURNING id, key
        `,
			project.ID,
			authorIDs[issue.Fields.Creator.DisplayName],
			authorIDs[issue.Fields.Assignee.DisplayName],
			issue.Key,
			issue.Fields.Summary,
			issue.Fields.Description,
			issue.Fields.IssueType.Name,
			issue.Fields.Priority.Name,
			issue.Fields.Status.Name,
			issue.Fields.Created.Time,
			issue.Fields.Closed.Time,
			issue.Fields.Updated.Time,
			issue.Fields.Timetracking.TimeSpentSeconds,
		)

		for _, history := range issue.Changelogs.Histories {
			for _, item := range history.Items {
				if item.Field == "status" {
					statusChanges = append(statusChanges, StatusChangeData{
						AuthorName: history.Author.DisplayName,
						ChangeTime: history.Created.Time,
						FromStatus: item.FromString,
						ToStatus:   item.ToString,
						IssueKey:   issue.Key,
					})
				}
			}
		}
	}

	br := tx.SendBatch(ctx, issueBatch)
	defer br.Close()

	for range project.Issues {
		var id int
		var key string
		if err := br.QueryRow().Scan(&id, &key); err != nil {
			return fmt.Errorf("failed to save issue: %w", err)
		}
		issueKeyToID[key] = id
	}
	if err := br.Close(); err != nil {
		return fmt.Errorf("failed to close issue batch: %w", err)
	}

	if len(statusChanges) > 0 {
		statusChangeBatch := &pgx.Batch{}

		for _, sc := range statusChanges {
			statusChangeBatch.Queue(`
                INSERT INTO StatusChanges (issueId, authorId, changeTime, fromStatus, toStatus)
                VALUES ($1, $2, $3, $4, $5)
                ON CONFLICT DO NOTHING
            `,
				issueKeyToID[sc.IssueKey],
				authorIDs[sc.AuthorName],
				sc.ChangeTime,
				sc.FromStatus,
				sc.ToStatus,
			)
		}

		sr := tx.SendBatch(ctx, statusChangeBatch)
		if err := sr.Close(); err != nil {
			return fmt.Errorf("failed to save status changes: %w", err)
		}
	}

	return tx.Commit(ctx)
}

type StatusChangeData struct {
	IssueKey   string
	AuthorName string
	ChangeTime time.Time
	FromStatus string
	ToStatus   string
}

func (p *ProjectRepository) Close() error {
	p.db.Close()
	return nil
}
