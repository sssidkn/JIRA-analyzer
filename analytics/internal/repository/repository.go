package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/sssidkn/JIRA-analyzer/internal/dto"
)

type Repository interface {
	//TODO gets for all tasks
	MakeTaskOne(ctx context.Context, key string) (*[]dto.IssueTaskOne, error)
	MakeTaskTwo(ctx context.Context, key string) (*[]dto.IssueTaskTwo, error)
}

type repo struct {
	db *pgx.Conn
}

func New(db *pgx.Conn) *repo {
	return &repo{db: db}
}

func (r *repo) MakeTaskOne(ctx context.Context, key string) (*[]dto.IssueTaskOne, error) {
	exist, err := r.checkExistenceOfProject(key)
	if err != nil {
		return nil, ErrExistence(err)
	}
	if !exist {
		return nil, ErrNotExist
	}

	var issues []dto.IssueTaskOne
	query := `WITH time_categories AS (
	  SELECT i.id,
		CASE 
		  WHEN i.timespent <= 3600 THEN '1 hour'
		  WHEN i.timespent <= 18000 THEN '1-5 hours'
		  WHEN i.timespent <= 36000 THEN '5-10 hours'
		  WHEN i.timespent <= 72000 THEN '10-20 hours'
		  WHEN i.timespent <= 86400 THEN '20-24 hours'
		  WHEN i.timespent <= 172800 THEN '1-2 days'
		  WHEN i.timespent <= 432000 THEN '2-5 days'
		  WHEN i.timespent <= 864000 THEN '5-10 days'
		  WHEN i.timespent <= 1296000 THEN '10-15 days'
 		  WHEN i.timespent <= 1728000 THEN '15-20 days'
          WHEN i.timespent <= 2592000 THEN '20-30 days' 
		  ELSE 'more'
		END AS time_category,
		CASE 
		  WHEN i.timespent <= 3600 THEN 1
		  WHEN i.timespent <= 18000 THEN 2
		  WHEN i.timespent <= 36000 THEN 3
		  WHEN i.timespent <= 72000 THEN 4
		  WHEN i.timespent <= 86400 THEN 5
		  WHEN i.timespent <= 172800 THEN 6
		  WHEN i.timespent <= 432000 THEN 7
		  WHEN i.timespent <= 864000 THEN 8
          WHEN i.timespent <= 1296000 THEN 9
		  WHEN i.timespent <= 1728000 THEN 10
          WHEN i.timespent <= 2592000 THEN 11
		  ELSE 12
		END AS category_order
		  FROM issue i
		  INNER JOIN projects p ON p.id = i.projectid
		  WHERE p.key = $1
		  AND i.status IN ('Closed', 'Resolved')
          AND i.timespent IS NOT NULL
		)
		SELECT 
		  time_category,
		  COUNT(id) AS task_count,
		  category_order
		FROM time_categories
		GROUP BY time_category, category_order
		ORDER BY category_order`

	rows, err := r.db.Query(ctx, query, key)
	if err != nil {
		return nil, ErrSelect(err)
	}
	defer rows.Close()

	for rows.Next() {
		var issue dto.IssueTaskOne
		var category int
		err = rows.Scan(&issue.Time, &issue.Count, &category)
		issues = append(issues, issue)
	}
	return &issues, nil
}

func (r *repo) MakeTaskTwo(ctx context.Context, key string) (*[]dto.IssueTaskTwo, error) {
	exist, err := r.checkExistenceOfProject(key)
	if err != nil {
		return nil, ErrExistence(err)
	}
	if !exist {
		return nil, ErrNotExist
	}

	var issues []dto.IssueTaskTwo
	query := `WITH priority_categories AS (
		SELECT i.id, 
		CASE
		WHEN i.priority = 'Blocker' THEN 'blocker'
		WHEN i.priority = 'Critical' THEN 'critical'
		WHEN i.priority = 'Major' THEN 'major'
		WHEN i.priority = 'Minor' THEN 'minor'
		WHEN i.priority = 'Trivial' THEN 'trivial'
		END AS priority_category,
		CASE
		WHEN i.priority = 'Blocker' THEN 1
		WHEN i.priority = 'Critical' THEN 2
		WHEN i.priority = 'Major' THEN 3
		WHEN i.priority = 'Minor' THEN 4
		WHEN i.priority = 'Trivial' THEN 5
		END AS priority_order
		FROM issue i
		  INNER JOIN projects p ON p.id = i.projectid
		  WHERE p.key = $1
		)
		SELECT 
		  priority_category,
		  COUNT(id) AS task_count,
		  priority_order
		FROM priority_categories
		GROUP BY priority_category, priority_order
		ORDER BY priority_order`

	rows, err := r.db.Query(ctx, query, key)
	if err != nil {
		return nil, ErrSelect(err)
	}
	defer rows.Close()

	for rows.Next() {
		var issue dto.IssueTaskTwo
		var category int
		err = rows.Scan(&issue.Priority, &issue.Count, &category)
		issues = append(issues, issue)
	}
	return &issues, nil
}

func (r *repo) checkExistenceOfProject(key string) (bool, error) {
	var exist bool
	err := r.db.QueryRow(context.Background(), `SELECT EXISTS(SELECT 1 FROM projects WHERE key = $1 )`, key).Scan(&exist)
	return exist, err
}
