package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/sssidkn/JIRA-analyzer/internal/dto"
)

type Repository interface {
	MakeTaskOne(ctx context.Context, key string) (*[]dto.IssueTaskOne, error)
	MakeTaskTwo(ctx context.Context, key string) (*[]dto.IssueTaskTwo, error)
	GetTaskOne(ctx context.Context, key string) (*[]dto.IssueTaskOne, error)
	GetTaskTwo(ctx context.Context, key string) (*[]dto.IssueTaskTwo, error)
	DeleteTasks(ctx context.Context, key string) (bool, error)
	IsAnalyzed(ctx context.Context, key string) (bool, error)
}

type repo struct {
	db *pgx.Conn
}

func New(db *pgx.Conn) *repo {
	return &repo{db: db}
}

func (r *repo) MakeTaskOne(ctx context.Context, key string) (*[]dto.IssueTaskOne, error) {
	id, err := r.checkExistenceOfProject(key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotExistProject
		}
		return nil, ErrExistence(err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, ErrBeginTransaction(err)
	}
	defer tx.Rollback(ctx)

	var exist bool
	query := `SELECT EXISTS (SELECT 1 FROM opentasktime WHERE projectid = $1)`
	err = tx.QueryRow(ctx, query, id).Scan(&exist)
	if err != nil {
		return nil, ErrExistence(err)
	}
	if exist {
		return nil, ErrAlreadyExist
	}

	var issues []dto.IssueTaskOne
	query = `WITH time_categories AS (
	  SELECT id,
		CASE 
		  WHEN timespent <= 3600 THEN '1 hour'
		  WHEN timespent <= 18000 THEN '1-5 hours'
		  WHEN timespent <= 36000 THEN '5-10 hours'
		  WHEN timespent <= 72000 THEN '10-20 hours'
		  WHEN timespent <= 86400 THEN '20-24 hours'
		  WHEN timespent <= 172800 THEN '1-2 days'
		  WHEN timespent <= 432000 THEN '2-5 days'
		  WHEN timespent <= 864000 THEN '5-10 days'
		  WHEN timespent <= 1296000 THEN '10-15 days'
 		  WHEN timespent <= 1728000 THEN '15-20 days'
          WHEN timespent <= 2592000 THEN '20-30 days' 
		  ELSE 'more'
		END AS time_category,
		CASE 
		  WHEN timespent <= 3600 THEN 1
		  WHEN timespent <= 18000 THEN 2
		  WHEN timespent <= 36000 THEN 3
		  WHEN timespent <= 72000 THEN 4
		  WHEN timespent <= 86400 THEN 5
		  WHEN timespent <= 172800 THEN 6
		  WHEN timespent <= 432000 THEN 7
		  WHEN timespent <= 864000 THEN 8
          WHEN timespent <= 1296000 THEN 9
		  WHEN timespent <= 1728000 THEN 10
          WHEN timespent <= 2592000 THEN 11
		  ELSE 12
		END AS category_order
		  FROM issue 
		  WHERE projectid = $1
		  AND status IN ('Closed', 'Resolved')
          AND timespent IS NOT NULL
		)
		SELECT 
		  time_category,
		  COUNT(id) AS task_count,
		  category_order
		FROM time_categories
		GROUP BY time_category, category_order
		ORDER BY category_order`

	rows, err := tx.Query(ctx, query, id)
	if err != nil {
		return nil, ErrSelect(err)
	}
	defer rows.Close()

	for rows.Next() {
		var issue dto.IssueTaskOne
		var category int
		err = rows.Scan(&issue.Time, &issue.Count, &category)
		if err != nil {
			return nil, ErrScan(err)
		}
		issues = append(issues, issue)
	}

	query = `INSERT INTO opentasktime (projectid, data) VALUES ($1, $2)`
	_, err = tx.Exec(ctx, query, id, issues)
	if err != nil {
		return nil, ErrInsert(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, ErrCommitTransaction(err)
	}
	return &issues, nil
}

func (r *repo) MakeTaskTwo(ctx context.Context, key string) (*[]dto.IssueTaskTwo, error) {
	id, err := r.checkExistenceOfProject(key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotExistProject
		}
		return nil, ErrExistence(err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, ErrBeginTransaction(err)
	}
	defer tx.Rollback(ctx)

	var exist bool
	query := `SELECT EXISTS (SELECT 1 FROM taskprioritycount WHERE projectid = $1)`
	err = tx.QueryRow(ctx, query, id).Scan(&exist)
	if err != nil {
		return nil, ErrExistence(err)
	}
	if exist {
		return nil, ErrAlreadyExist
	}

	var issues []dto.IssueTaskTwo
	query = `WITH priority_categories AS (
		SELECT id, 
		CASE
		WHEN priority = 'Blocker' THEN 'blocker'
		WHEN priority = 'Critical' THEN 'critical'
		WHEN priority = 'Major' THEN 'major'
		WHEN priority = 'Minor' THEN 'minor'
		WHEN priority = 'Trivial' THEN 'trivial'
		END AS priority_category,
		CASE
		WHEN priority = 'Blocker' THEN 1
		WHEN priority = 'Critical' THEN 2
		WHEN priority = 'Major' THEN 3
		WHEN priority = 'Minor' THEN 4
		WHEN priority = 'Trivial' THEN 5
		END AS priority_order
		FROM issue 
		   WHERE projectid = $1
		)
		SELECT 
		  priority_category,
		  COUNT(id) AS task_count,
		  priority_order
		FROM priority_categories
		GROUP BY priority_category, priority_order
		ORDER BY priority_order`

	rows, err := tx.Query(ctx, query, id)
	if err != nil {
		return nil, ErrSelect(err)
	}
	defer rows.Close()

	for rows.Next() {
		var issue dto.IssueTaskTwo
		var category int
		err = rows.Scan(&issue.Priority, &issue.Count, &category)
		if err != nil {
			return nil, ErrScan(err)
		}
		issues = append(issues, issue)
	}

	query = `INSERT INTO taskprioritycount (projectid, data) VALUES ($1, $2)`
	_, err = tx.Exec(ctx, query, id, issues)
	if err != nil {
		return nil, ErrInsert(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, ErrCommitTransaction(err)
	}

	return &issues, nil
}

func (r *repo) GetTaskOne(ctx context.Context, key string) (*[]dto.IssueTaskOne, error) {
	id, err := r.checkExistenceOfProject(key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotExistProject
		}
		return nil, ErrExistence(err)
	}

	var issues []dto.IssueTaskOne
	query := `SELECT data FROM opentasktime WHERE projectid = $1`
	err = r.db.QueryRow(ctx, query, id).Scan(&issues)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotExistData
		}
		return nil, ErrScan(err)
	}
	return &issues, nil
}

func (r *repo) GetTaskTwo(ctx context.Context, key string) (*[]dto.IssueTaskTwo, error) {
	id, err := r.checkExistenceOfProject(key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotExistProject
		}
		return nil, ErrExistence(err)
	}

	var issues []dto.IssueTaskTwo
	query := `SELECT data FROM taskprioritycount WHERE projectid = $1`
	err = r.db.QueryRow(ctx, query, id).Scan(&issues)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotExistData
		}
		return nil, ErrScan(err)
	}
	return &issues, nil
}

func (r *repo) DeleteTasks(ctx context.Context, key string) (bool, error) {
	id, err := r.checkExistenceOfProject(key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrNotExistProject
		}
		return false, ErrExistence(err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, ErrBeginTransaction(err)
	}
	defer tx.Rollback(ctx)

	query := `DELETE FROM opentasktime WHERE projectid = $1`
	res1, err := tx.Exec(ctx, query, id)
	if err != nil {
		return false, ErrDelete(err)
	}

	query = `DELETE FROM taskprioritycount WHERE projectid = $1`
	res2, err := tx.Exec(ctx, query, id)
	if err != nil {
		return false, ErrDelete(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return false, ErrCommitTransaction(err)
	}

	if res1.RowsAffected() == 0 && res2.RowsAffected() == 0 {
		return false, nil
	}

	return true, nil
}

func (r *repo) IsAnalyzed(ctx context.Context, key string) (bool, error) {
	id, err := r.checkExistenceOfProject(key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrNotExistProject
		}
		return false, ErrExistence(err)
	}

	var exist bool
	query := `SELECT EXISTS (SELECT 1 FROM opentasktime WHERE projectid = $1)`
	err = r.db.QueryRow(ctx, query, id).Scan(&exist)
	if err != nil {
		return false, ErrExistence(err)
	}
	if exist {
		return true, nil
	}

	query = `SELECT EXISTS (SELECT 1 FROM taskprioritycount WHERE projectid = $1)`
	err = r.db.QueryRow(ctx, query, id).Scan(&exist)
	if err != nil {
		return false, ErrExistence(err)
	}
	if exist {
		return true, nil
	}
	return false, nil
}

func (r *repo) checkExistenceOfProject(key string) (int, error) {
	var id int
	err := r.db.QueryRow(context.Background(), `SELECT id FROM projects WHERE key = $1`, key).Scan(&id)
	return id, err
}
