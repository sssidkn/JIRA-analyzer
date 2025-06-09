package models

import (
	"time"
)

type JiraProject struct {
	ID              string      `json:"id"`
	Key             string      `json:"key"`
	Name            string      `json:"name"`
	Issues          []JiraIssue `json:"issues"`
	TotalIssueCount int         `json:"totalIssuesCount"`
	LastUpdate      time.Time
}

type ProjectInfo struct {
	ID         string `json:"id"`
	Key        string `json:"key"`
	Name       string `json:"name"`
	LastUpdate time.Time
}
