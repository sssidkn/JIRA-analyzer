package models

type JiraProject struct {
	ID     string      `json:"id"`
	Key    string      `json:"key"`
	Name   string      `json:"name"`
	Issues []JiraIssue `json:"issues"`
}

type ProjectInfo struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

type JiraUser struct {
	AccountID   string `json:"accountId"`
	DisplayName string `json:"displayName"`
}

type JiraIssue struct {
	ID         string    `json:"id"`
	Key        string    `json:"key"`
	Fields     Fields    `json:"fields"`
	Changelogs Changelog `json:"changelog"`
}

type Fields struct {
	Summary     string `json:"summary"`
	Description string `json:"description"`
	IssueType   struct {
		Name string `json:"name"`
	} `json:"issuetype"`
	Priority struct {
		Name string `json:"name"`
	} `json:"priority"`
	Status struct {
		Name string `json:"name"`
	} `json:"status"`
	Creator    JiraUser `json:"creator"`
	Assignee   JiraUser `json:"assignee"`
	Created    string   `json:"created"`
	Updated    string   `json:"updated"`
	Resolution struct {
		Date string `json:"resolutiondate"`
	} `json:"resolution"`
	Timespent *int `json:"timespent"`
}

type Changelog struct {
	Histories []History `json:"histories"`
}

type History struct {
	Created string   `json:"created"`
	Author  JiraUser `json:"author"`
	Items   []Item   `json:"items"`
}

type Item struct {
	Field      string `json:"field"`
	FromString string `json:"fromString"`
	ToString   string `json:"toString"`
}
