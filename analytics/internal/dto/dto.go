package dto

type IssueTaskOne struct {
	Count int    `json:"count"`
	Time  string `json:"time"`
}

type IssueTaskTwo struct {
	Count    int    `json:"count"`
	Priority string `json:"priority"`
}
