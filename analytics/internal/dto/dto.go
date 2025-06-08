package dto

type IssueTaskOne struct {
	Count int    `json:"count"`
	Time  string `json:"time"`
}

type IssueTaskTwo struct {
	Count    int    `json:"count"`
	Priority string `json:"priority"`
}

type ComparisonTaskOne struct {
	Key  string         `json:"key"`
	Data []IssueTaskOne `json:"data"`
}

type ComparisonTaskTwo struct {
	Key  string         `json:"key"`
	Data []IssueTaskTwo `json:"data"`
}
