package repository

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
	"github.com/sssidkn/JIRA-analyzer/internal/models"
)

type Repository interface {
	GetProjects(ctx context.Context, limit int, offset int) (*[]models.Project, int, error)
	GetProject(ctx context.Context, id int) (*models.ProjectInfo, error)
	DeleteProject(ctx context.Context, id int) error
	GetIssue(ctx context.Context, id int) (*models.IssueInfo, error)
	GetIssuesByProject(ctx context.Context, projectId int, limit int, offset int) (*[]models.Issue, int, error)
	GetHistoryByIssue(ctx context.Context, issueId int) (*[]models.History, error)
	GetHistoryByAuthor(ctx context.Context, authorId int) (*[]models.History, error)
}

type repo struct {
	db *pgx.Conn
}

func New(db *pgx.Conn) *repo {
	return &repo{db: db}
}

func (r *repo) GetProjects(ctx context.Context, limit int, offset int) (*[]models.Project, int, error) {
	var projects []models.Project
	query := `SELECT id, key, title from projects LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, ErrSelect(err)
	}
	defer rows.Close()

	for rows.Next() {
		var p models.Project
		err = rows.Scan(&p.Id, &p.Key, &p.Name)
		if err != nil {
			return nil, 0, ErrScan(err)
		}
		projects = append(projects, p)
	}

	var total int
	query = `SELECT COUNT(*) FROM projects`
	err = r.db.QueryRow(ctx, query).Scan(&total)
	if err != nil {
		return nil, 0, ErrSelect(err)
	}
	return &projects, total, nil
}

func (r *repo) GetProject(ctx context.Context, id int) (*models.ProjectInfo, error) {
	exist, err := r.checkExistenceOfProject(id)
	if err != nil {
		return nil, ErrExistence(err)
	}
	if !exist {
		return nil, ErrNotExist
	}

	var project models.ProjectInfo
	query := `SELECT id, key, title from projects WHERE id = $1`
	err = r.db.QueryRow(ctx, query, id).Scan(&project.Id, &project.Key, &project.Name)
	if err != nil {
		return nil, ErrScan(err)
	}

	query = `SELECT COUNT(*) FROM issue WHERE projectid = $1`
	err = r.db.QueryRow(ctx, query, id).Scan(&project.AllIssuesCount)
	if err != nil {
		return nil, ErrScan(err)
	}

	query = `SELECT COUNT(*) FROM issue WHERE projectid = $1 AND status = $2`
	err = r.db.QueryRow(ctx, query, id, "Opened").Scan(&project.OpenedIssuesCount)
	if err != nil {
		return nil, ErrScan(err)
	}
	err = r.db.QueryRow(ctx, query, id, "Closed").Scan(&project.ClosedIssuesCount)
	if err != nil {
		return nil, ErrScan(err)
	}
	err = r.db.QueryRow(ctx, query, id, "Resolved").Scan(&project.ResolvedIssuesCount)
	if err != nil {
		return nil, ErrScan(err)
	}
	err = r.db.QueryRow(ctx, query, id, "Reopened").Scan(&project.ReopenedIssuesCount)
	if err != nil {
		return nil, ErrScan(err)
	}
	err = r.db.QueryRow(ctx, query, id, "In Progress").Scan(&project.ProgressIssuesCount)
	if err != nil {
		return nil, ErrScan(err)
	}

	var avgTime sql.NullFloat64
	query = `SELECT AVG(timespent) FROM issue WHERE projectid = $1`
	err = r.db.QueryRow(ctx, query, id).Scan(&avgTime)
	if err != nil {
		return nil, ErrScan(err)
	}
	if avgTime.Valid {
		project.AverageTime = avgTime.Float64
	} else {
		project.AverageTime = 0
	}

	query = `SELECT 
    	COUNT(*) / 7.0 
		FROM 
    		issue
		WHERE 
		projectid = $1
		AND
    	createdtime >= CURRENT_DATE - INTERVAL '7 days'`
	err = r.db.QueryRow(ctx, query, id).Scan(&project.AverageIssuesCount)
	if err != nil {
		return nil, ErrScan(err)
	}
	return &project, nil
}

func (r *repo) DeleteProject(ctx context.Context, id int) error {
	exist, err := r.checkExistenceOfProject(id)
	if err != nil {
		return ErrExistence(err)
	}
	if !exist {
		return ErrNotExist
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return ErrBeginTransaction(err)
	}
	defer tx.Rollback(ctx)

	var issuesIds []int
	query := `SELECT id FROM issue WHERE projectid = $1`
	rows, err := tx.Query(ctx, query, id)
	if err != nil {
		return ErrSelect(err)
	}
	defer rows.Close()
	for rows.Next() {
		var issueId int
		err = rows.Scan(&issueId)
		if err != nil {
			return ErrScan(err)
		}
		issuesIds = append(issuesIds, issueId)
	}

	query = `DELETE FROM statuschanges WHERE issueid = $1`
	for _, issueId := range issuesIds {
		_, err = tx.Exec(ctx, query, issueId)
		if err != nil {
			return ErrDelete(err)
		}
	}

	query = `DELETE FROM issue WHERE projectid = $1`
	_, err = tx.Exec(ctx, query, id)
	if err != nil {
		return ErrDelete(err)
	}

	query = `DELETE FROM projects WHERE id = $1`
	_, err = tx.Exec(ctx, query, id)
	if err != nil {
		return ErrDelete(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return ErrCommitTransaction(err)
	}
	return nil
}

func (r *repo) GetIssue(ctx context.Context, id int) (*models.IssueInfo, error) {
	exist, err := r.checkExistenceOfIssue(id)
	if err != nil {
		return nil, ErrExistence(err)
	}
	if !exist {
		return nil, ErrNotExist
	}

	var issue models.IssueInfo
	var timeSpent sql.NullInt32
	query := `SELECT * from issue WHERE id = $1`
	err = r.db.QueryRow(ctx, query, id).Scan(&issue.Id, &issue.ProjectId, &issue.AuthorId, &issue.AssigneeId, &issue.Key, &issue.Summary,
		&issue.Description, &issue.Type, &issue.Priority, &issue.Status, &issue.CreatedTime, &issue.ClosedTime, &issue.UpdatedTime, &timeSpent)
	if err != nil {
		return nil, ErrScan(err)
	}
	if timeSpent.Valid {
		issue.TimeSpent = timeSpent.Int32
	}

	query = `SELECT name FROM author WHERE id = $1`
	err = r.db.QueryRow(ctx, query, id).Scan(&issue.AuthorName)
	if err != nil {
		return nil, ErrScan(err)
	}

	query = `SELECT COUNT(*) FROM statuschanges WHERE issueid = $1`
	err = r.db.QueryRow(ctx, query, id).Scan(&issue.ChangeStatusCount)
	if err != nil {
		return nil, ErrScan(err)
	}
	return &issue, nil
}

func (r *repo) GetIssuesByProject(ctx context.Context, projectId int, limit int, offset int) (*[]models.Issue, int, error) {
	exist, err := r.checkExistenceOfProject(projectId)
	if err != nil {
		return nil, 0, ErrExistence(err)
	}
	if !exist {
		return nil, 0, ErrNotExist
	}

	var issues []models.Issue
	query := `SELECT id, projectid, authorid, type from issue WHERE projectid = $1 LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, projectId, limit, offset)
	if err != nil {
		return nil, 0, ErrSelect(err)
	}
	defer rows.Close()

	for rows.Next() {
		var issue models.Issue
		err = rows.Scan(&issue.Id, &issue.ProjectId, &issue.AuthorId, &issue.Type)
		if err != nil {
			return nil, 0, ErrScan(err)
		}
		issues = append(issues, issue)
	}

	var total int
	query = `SELECT COUNT(*) FROM issue WHERE projectid = $1`
	err = r.db.QueryRow(ctx, query, projectId).Scan(&total)
	if err != nil {
		return nil, 0, ErrScan(err)
	}

	return &issues, total, nil
}

func (r *repo) GetHistoryByIssue(ctx context.Context, issueId int) (*[]models.History, error) {
	exist, err := r.checkExistenceOfIssue(issueId)
	if err != nil {
		return nil, ErrExistence(err)
	}
	if !exist {
		return nil, ErrNotExist
	}

	var histories []models.History
	query := `SELECT * from statuschanges WHERE issueid = $1`
	rows, err := r.db.Query(ctx, query, issueId)
	if err != nil {
		return nil, ErrSelect(err)
	}
	defer rows.Close()

	for rows.Next() {
		var history models.History
		err = rows.Scan(&history.IssueId, &history.AuthorId, &history.ChangeTime, &history.FromStatus, &history.ToStatus)
		if err != nil {
			return nil, ErrScan(err)
		}
		histories = append(histories, history)
	}
	return &histories, nil
}

func (r *repo) GetHistoryByAuthor(ctx context.Context, authorId int) (*[]models.History, error) {
	exist, err := r.checkExistenceOfAuthor(authorId)
	if err != nil {
		return nil, ErrExistence(err)
	}
	if !exist {
		return nil, ErrNotExist
	}

	var histories []models.History
	query := `SELECT * from statuschanges WHERE authorid = $1`
	rows, err := r.db.Query(ctx, query, authorId)
	if err != nil {
		return nil, ErrSelect(err)
	}
	defer rows.Close()

	for rows.Next() {
		var history models.History
		err = rows.Scan(&history.IssueId, &history.AuthorId, &history.ChangeTime, &history.FromStatus, &history.ToStatus)
		if err != nil {
			return nil, ErrScan(err)
		}
		histories = append(histories, history)
	}
	return &histories, nil
}

func (r *repo) checkExistenceOfProject(id int) (bool, error) {
	var exist bool
	err := r.db.QueryRow(context.Background(), `SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 )`, id).Scan(&exist)
	return exist, err
}

func (r *repo) checkExistenceOfIssue(id int) (bool, error) {
	var exist bool
	err := r.db.QueryRow(context.Background(), `SELECT EXISTS(SELECT 1 FROM issue WHERE id = $1 )`, id).Scan(&exist)
	return exist, err
}

func (r *repo) checkExistenceOfAuthor(id int) (bool, error) {
	var exist bool
	err := r.db.QueryRow(context.Background(), `SELECT EXISTS(SELECT 1 FROM author WHERE id = $1 )`, id).Scan(&exist)
	return exist, err
}
