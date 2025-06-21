package models

import (
	"strings"
	"time"
)

type JiraTime struct {
	time.Time
}

func (jt *JiraTime) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" {
		return nil
	}
	formats := []string{
		"2006-01-02T15:04:05.000-0700",
		"2006-01-02T15:04:05.000+0000",
		time.RFC3339,
	}

	s = strings.Trim(s, `"`)
	var err error

	for _, format := range formats {
		jt.Time, err = time.Parse(format, s)
		if err == nil {
			return nil
		}
	}

	return nil
}
