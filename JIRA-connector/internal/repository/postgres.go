package repository

import (
	"context"
	"errors"
	"jira-connector/internal/models"
	"jira-connector/pkg/logger"

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

// TODO: Задачка со звездочкой
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
	).Scan(&pi.Key, &pi.Key, &pi.Name, &pi.LastUpdate)
	if err != nil {
		return nil, err
	}
	return pi, nil
}

func (p *ProjectRepository) SaveProject(ctx context.Context, project Project) error {
	tx, err := p.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	issues := project.Issues
	_, err = tx.Prepare(ctx, "save-project", `
		INSERT INTO Projects (title, key) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE
		                                                                  SET lastUpdate = $3
		                                                                  RETURNING id
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
		err = tx.QueryRow(ctx, "save-project", project.Name, project.Key, project.LastUpdate).Scan(&projectID)
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
		createdTime := issue.Fields.Created.Time

		updatedTime := issue.Fields.Updated.Time

		closedTime := issue.Fields.Closed.Time
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
			issue.Fields.Timetracking.TimeSpentSeconds,
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
			changeTime := history.Created.Time
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
