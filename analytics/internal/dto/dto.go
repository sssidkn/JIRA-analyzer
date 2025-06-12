package dto

// IssueTaskOne represents task one data
type IssueTaskOne struct {
	Count int    `json:"count"`
	Time  string `json:"time"`
}

// IssueTaskTwo represents task two data
type IssueTaskTwo struct {
	Count    int    `json:"count"`
	Priority string `json:"priority"`
}

// ComparisonTaskOne represents comparison data for task one
type ComparisonTaskOne struct {
	Key  string         `json:"key"`
	Data []IssueTaskOne `json:"data"`
}

// ComparisonTaskTwo represents comparison data for task two
type ComparisonTaskTwo struct {
	Key  string         `json:"key"`
	Data []IssueTaskTwo `json:"data"`
}
