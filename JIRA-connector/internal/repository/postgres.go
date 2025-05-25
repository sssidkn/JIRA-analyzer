package repository

import (
	"context"
	"errors"
	"jira-connector/internal/models"
	"jira-connector/pkg/logger"
	"strings"
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

func parseJiraTime(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, nil
	}
	normalized := strings.Replace(timeStr, "+0000", "Z", 1)
	return time.Parse(time.RFC3339, normalized)
}

func (p *ProjectRepository) ProjectExists(ctx context.Context, projectKey string) (bool, error) {
	var exists bool
	err := p.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM Projects WHERE key = $1)",
		projectKey,
	).Scan(&exists)

	if err != nil {
		return false, err
	}
	return exists, nil
}

func (p *ProjectRepository) SaveProject(ctx context.Context, project Project) error {
	tx, err := p.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	issues := project.Issues
	_, err = tx.Prepare(ctx, "save-project", `
		INSERT INTO Projects (title, key, lastUpdate) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING RETURNING id
	`)
	if err != nil {
		return err
	}

	_, err = tx.Prepare(ctx, "save-author", `
		INSERT INTO Author (name) VALUES ($1) ON CONFLICT DO NOTHING RETURNING id
	`)
	if err != nil {
		return err
	}

	_, err = tx.Prepare(ctx, "save-issue", `
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
		RETURNING id
	`)
	if err != nil {
		return err
	}

	_, err = tx.Prepare(ctx, "save-status-change", `
		INSERT INTO StatusChanges (issueId, authorId, changeTime, fromStatus, toStatus)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT DO NOTHING 
	`)
	if err != nil {
		return err
	}

	for _, issue := range issues {
		var projectID int
		lastUpdate, err := parseJiraTime(issue.Fields.Updated)
		err = tx.QueryRow(ctx, "save-project", project.Name, project.Key, lastUpdate).Scan(&projectID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		if errors.Is(err, pgx.ErrNoRows) {
			err = tx.QueryRow(ctx, `SELECT id FROM Projects WHERE key = $1`, project.Key).Scan(&projectID)
			if err != nil {
				return err
			}
		}

		var authorID int
		err = tx.QueryRow(ctx, "save-author", issue.Fields.Creator.DisplayName).Scan(&authorID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		if errors.Is(err, pgx.ErrNoRows) {
			err = tx.QueryRow(ctx, `SELECT id FROM Author WHERE name = $1`, issue.Fields.Creator.DisplayName).Scan(&authorID)
			if err != nil {
				return err
			}
		}

		var assigneeID int
		if issue.Fields.Assignee.DisplayName != "" {
			err = tx.QueryRow(ctx, "save-author", issue.Fields.Assignee.DisplayName).Scan(&assigneeID)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return err
			}

			if errors.Is(err, pgx.ErrNoRows) {
				err = tx.QueryRow(ctx, `SELECT id FROM Author WHERE name = $1`, issue.Fields.Assignee.DisplayName).Scan(&assigneeID)
				if err != nil {
					return err
				}
			}
		}
		createdTime, err := parseJiraTime(issue.Fields.Created)
		if err != nil {
			return err
		}

		updatedTime, err := parseJiraTime(issue.Fields.Updated)
		if err != nil {
			return err
		}

		var closedTime time.Time
		if issue.Fields.Resolution.Date != "" {
			closedTime, err = parseJiraTime(issue.Fields.Resolution.Date)
			if err != nil {
				return err
			}
		}
		var issueID int
		err = tx.QueryRow(ctx, "save-issue",
			projectID,
			authorID,
			assigneeID,
			issue.Key,
			issue.Fields.Summary,
			issue.Fields.Description,
			issue.Fields.IssueType.Name,
			issue.Fields.Priority.Name,
			issue.Fields.Status.Name,
			createdTime,
			closedTime,
			updatedTime,
			issue.Fields.Timespent,
		).Scan(&issueID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		if errors.Is(err, pgx.ErrNoRows) {
			err = tx.QueryRow(ctx, `SELECT id FROM Issue WHERE key = $1`, issue.Key).Scan(&issueID)
			if err != nil {
				return err
			}
		}
		changelog := issue.Changelogs
		for _, history := range changelog.Histories {
			changeTime, err := parseJiraTime(history.Created)
			if err != nil {
				return err
			}

			for _, item := range history.Items {
				if item.Field == "status" {
					var statusChangeAuthorID int
					err = tx.QueryRow(ctx, "save-author", history.Author.DisplayName).Scan(&statusChangeAuthorID)
					if err != nil && !errors.Is(err, pgx.ErrNoRows) {
						return err
					}

					if errors.Is(err, pgx.ErrNoRows) {
						err = tx.QueryRow(ctx, `SELECT id FROM Author WHERE name = $1`, history.Author.DisplayName).Scan(&statusChangeAuthorID)
						if err != nil {
							return err
						}
					}

					_, err = tx.Exec(ctx, "save-status-change",
						issueID,
						statusChangeAuthorID,
						changeTime,
						item.FromString,
						item.ToString,
					)
					if err != nil {
						return err
					}
				}

			}
		}
	}

	return tx.Commit(ctx)
}

func (p *ProjectRepository) Close() error {
	p.db.Close()
	return nil
}
