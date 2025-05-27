package models

import "time"

type Project struct {
	Id   int    `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

type ProjectInfo struct {
	Id                  int     `json:"id"`
	Key                 string  `json:"key"`
	Name                string  `json:"name"`
	AllIssuesCount      int     `json:"allIssuesCount"`
	OpenedIssuesCount   int     `json:"openedIssuesCount"`
	ClosedIssuesCount   int     `json:"closedIssuesCount"`
	ResolvedIssuesCount int     `json:"resolvedIssuesCount"`
	ReopenedIssuesCount int     `json:"reopenedIssuesCount"`
	ProgressIssuesCount int     `json:"progressIssuesCount"`
	AverageTime         float64 `json:"averageTime"`
	AverageIssuesCount  float64 `json:"averageIssuesCount"`
}

type Issue struct {
	Id        int    `json:"id"`
	ProjectId int    `json:"projectId"`
	AuthorId  int    `json:"authorId"`
	Type      string `json:"type"`
}

type IssueInfo struct {
	Id                int       `json:"id"`
	ProjectId         int       `json:"projectId"`
	AuthorId          int       `json:"authorId"`
	AuthorName        string    `json:"authorName"`
	AssigneeId        int       `json:"assignedId"`
	Key               string    `json:"key"`
	Summary           string    `json:"summary"`
	Description       string    `json:"description"`
	Type              string    `json:"type"`
	Priority          string    `json:"priority"`
	Status            string    `json:"status"`
	CreatedTime       time.Time `json:"created_time"`
	ClosedTime        time.Time `json:"closed_time"`
	UpdatedTime       time.Time `json:"updated_time"`
	TimeSpent         int32     `json:"timespent"`
	ChangeStatusCount int       `json:"change_status_count"`
}

type History struct {
	IssueId    int       `json:"issueId"`
	AuthorId   int       `json:"authorId"`
	ChangeTime time.Time `json:"changeTime"`
	FromStatus string    `json:"fromStatus"`
	ToStatus   string    `json:"toStatus"`
}

type Link struct {
	URL string `json:"href"`
}

type ReferencesLinks struct {
	LinkSelf      Link   `json:"self"`
	LinkIssues    []Link `json:"issues"`
	LinkProjects  []Link `json:"projects"`
	LinkHistories []Link `json:"histories"`
}

type Response struct {
	Links ReferencesLinks `json:"_links"`
	Data  interface{}     `json:"data"`
}

type Pagination struct {
	CurrentPage int `json:"currentPage"`
	PageCount   int `json:"pageCount"`
	Total       int `json:"total"`
}

type PaginatedResponse struct {
	Links    ReferencesLinks `json:"_links"`
	Data     interface{}     `json:"data"`
	PageInfo Pagination      `json:"pageInfo"`
}
