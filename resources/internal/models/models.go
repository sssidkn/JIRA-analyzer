package models

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
	AverageIssuesCount  int     `json:"averageIssuesCount"`
}

type Link struct {
	URL string `json:"href"`
}

type ReferencesLinks struct {
	LinkSelf      Link `json:"self"`
	LinkIssues    Link `json:"issues"`
	LinkProjects  Link `json:"projects"`
	LinkHistories Link `json:"histories"`
}

type Response struct {
	Links ReferencesLinks `json:"_links"`
	Data  interface{}     `json:"data"`
}
