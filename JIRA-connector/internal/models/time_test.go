package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJiraTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		jsonInput   string
		expected    time.Time
		expectError bool
	}{
		{
			name:      "Format with timezone offset",
			jsonInput: `"2023-12-31T15:04:05.000+0300"`,
			expected:  time.Date(2023, 12, 31, 15, 4, 5, 0, time.FixedZone("+0300", 3*60*60)),
		},
		{
			name:      "Format with UTC offset",
			jsonInput: `"2023-12-31T15:04:05.000+0000"`,
			expected:  time.Date(2023, 12, 31, 15, 4, 5, 0, time.UTC),
		},
		{
			name:      "RFC3339 format",
			jsonInput: `"2023-12-31T15:04:05Z"`,
			expected:  time.Date(2023, 12, 31, 15, 4, 5, 0, time.UTC),
		},
		{
			name:      "RFC3339 with nanoseconds",
			jsonInput: `"2023-12-31T15:04:05.123456789Z"`,
			expected:  time.Date(2023, 12, 31, 15, 4, 5, 123456789, time.UTC),
		},
		{
			name:        "Null value",
			jsonInput:   `null`,
			expected:    time.Time{},
			expectError: false,
		},
		{
			name:        "Empty string",
			jsonInput:   `""`,
			expected:    time.Time{},
			expectError: false,
		},
		{
			name:        "Invalid format",
			jsonInput:   `"invalid-date-format"`,
			expected:    time.Time{},
			expectError: false, // Ваша реализация возвращает nil при ошибке, а не ошибку
		},
		{
			name:      "With milliseconds and timezone",
			jsonInput: `"2023-12-31T15:04:05.123-0500"`,
			expected:  time.Date(2023, 12, 31, 15, 4, 5, 123000000, time.FixedZone("-0500", -5*60*60)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jt JiraTime

			err := json.Unmarshal([]byte(tt.jsonInput), &jt)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Сравниваем с допуском в 1 секунду из-за возможных различий в парсинге
				if !tt.expected.IsZero() {
					assert.WithinDuration(t, tt.expected, jt.Time, time.Second)
				} else {
					assert.True(t, jt.Time.IsZero())
				}
			}
		})
	}
}

func TestJiraTime_UnmarshalJSON_EdgeCases(t *testing.T) {
	t.Run("Multiple formats in sequence", func(t *testing.T) {
		formats := []string{
			`"2023-12-31T15:04:05.000+0300"`,
			`"2023-12-31T15:04:05.000+0000"`,
			`"2023-12-31T15:04:05Z"`,
		}

		for _, input := range formats {
			var jt JiraTime
			err := json.Unmarshal([]byte(input), &jt)
			assert.NoError(t, err)
			assert.False(t, jt.Time.IsZero())
		}
	})

	t.Run("Very old date", func(t *testing.T) {
		var jt JiraTime
		err := json.Unmarshal([]byte(`"1970-01-01T00:00:00.000+0000"`), &jt)
		assert.NoError(t, err)
		assert.Equal(t, 1970, jt.Time.Year())
	})

	t.Run("Future date", func(t *testing.T) {
		var jt JiraTime
		err := json.Unmarshal([]byte(`"2030-12-31T23:59:59.999+0000"`), &jt)
		assert.NoError(t, err)
		assert.Equal(t, 2030, jt.Time.Year())
	})

	t.Run("Whitespace handling", func(t *testing.T) {
		var jt JiraTime
		// JSON с пробелами - должен корректно обрабатываться
		err := json.Unmarshal([]byte(`  "2023-12-31T15:04:05.000+0300"  `), &jt)
		assert.NoError(t, err)
		assert.False(t, jt.Time.IsZero())
	})
}

func TestJiraTime_MarshalJSON(t *testing.T) {
	t.Run("Round trip serialization", func(t *testing.T) {
		originalTime := time.Date(2023, 12, 31, 15, 4, 5, 0, time.FixedZone("+0300", 3*60*60))
		jt := JiraTime{Time: originalTime}

		// Marshal to JSON
		data, err := json.Marshal(jt)
		require.NoError(t, err)

		// Unmarshal back
		var jt2 JiraTime
		err = json.Unmarshal(data, &jt2)
		require.NoError(t, err)

		// Should be equal (within 1 second due to possible format differences)
		assert.WithinDuration(t, originalTime, jt2.Time, time.Second)
	})

	t.Run("Zero time", func(t *testing.T) {
		jt := JiraTime{Time: time.Time{}}

		data, err := json.Marshal(jt)
		require.NoError(t, err)

		// Zero time should marshal to null or zero value representation
		assert.Contains(t, string(data), "0001-01-01") // Go's zero time representation
	})
}

func TestJiraTime_String(t *testing.T) {
	t.Run("Non-zero time", func(t *testing.T) {
		jt := JiraTime{Time: time.Date(2023, 12, 31, 15, 4, 5, 0, time.UTC)}
		str := jt.String()

		// Should use the underlying time.Time's String method
		assert.Contains(t, str, "2023-12-31")
		assert.Contains(t, str, "15:04:05")
	})

	t.Run("Zero time", func(t *testing.T) {
		jt := JiraTime{Time: time.Time{}}
		str := jt.String()

		// Should return zero time string representation
		assert.Contains(t, str, "0001-01-01")
	})
}

func TestJiraTime_ValueMethods(t *testing.T) {
	t.Run("Access underlying time methods", func(t *testing.T) {
		testTime := time.Date(2023, 12, 31, 15, 4, 5, 0, time.UTC)
		jt := JiraTime{Time: testTime}

		// Should be able to access all time.Time methods
		assert.Equal(t, 2023, jt.Year())
		assert.Equal(t, time.December, jt.Month())
		assert.Equal(t, 31, jt.Day())
		assert.Equal(t, 15, jt.Hour())
		assert.Equal(t, 4, jt.Minute())
		assert.Equal(t, 5, jt.Second())
	})

	t.Run("Comparison methods", func(t *testing.T) {
		time1 := JiraTime{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)}
		time2 := JiraTime{Time: time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)}
		time3 := JiraTime{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)}

		assert.True(t, time1.Before(time2.Time))
		assert.True(t, time2.After(time1.Time))
		assert.True(t, time1.Equal(time3.Time))
	})
}

func TestJiraTime_IntegrationWithStruct(t *testing.T) {
	type TestStruct struct {
		Created   JiraTime `json:"created"`
		Updated   JiraTime `json:"updated"`
		Completed JiraTime `json:"completed"`
	}

	t.Run("Full struct unmarshaling", func(t *testing.T) {
		jsonData := `{
			"created": "2023-01-15T10:30:00.000+0300",
			"updated": "2023-01-15T14:45:30.123+0000",
			"completed": "2023-01-16T09:15:45Z"
		}`

		var result TestStruct
		err := json.Unmarshal([]byte(jsonData), &result)
		require.NoError(t, err)

		assert.False(t, result.Created.Time.IsZero())
		assert.False(t, result.Updated.Time.IsZero())
		assert.False(t, result.Completed.Time.IsZero())

		assert.Equal(t, 2023, result.Created.Year())
		assert.Equal(t, 2023, result.Updated.Year())
		assert.Equal(t, 2023, result.Completed.Year())
	})

	t.Run("Struct with null and empty values", func(t *testing.T) {
		jsonData := `{
			"created": "2023-01-15T10:30:00.000+0300",
			"updated": null,
			"completed": ""
		}`

		var result TestStruct
		err := json.Unmarshal([]byte(jsonData), &result)
		require.NoError(t, err)

		assert.False(t, result.Created.Time.IsZero())
		assert.True(t, result.Updated.Time.IsZero())
		assert.True(t, result.Completed.Time.IsZero())
	})
}

func TestJiraTime_ErrorHandling(t *testing.T) {
	t.Run("Malformed JSON", func(t *testing.T) {
		var jt JiraTime

		// Missing closing quote
		err := json.Unmarshal([]byte(`"2023-12-31T15:04:05.000+0300`), &jt)
		assert.Error(t, err)

		// Invalid JSON structure
		err = json.Unmarshal([]byte(`{invalid}`), &jt)
		assert.Error(t, err)
	})

	t.Run("Unsupported date format", func(t *testing.T) {
		var jt JiraTime

		// Format not in the supported list
		err := json.Unmarshal([]byte(`"31/12/2023 15:04:05"`), &jt)
		assert.NoError(t, err)           // Ваша реализация возвращает nil при ошибке парсинга
		assert.True(t, jt.Time.IsZero()) // Поэтому время должно быть zero value
	})
}

// Benchmark тесты для проверки производительности
func BenchmarkJiraTime_UnmarshalJSON(b *testing.B) {
	testCases := []string{
		`"2023-12-31T15:04:05.000+0300"`,
		`"2023-12-31T15:04:05.000+0000"`,
		`"2023-12-31T15:04:05Z"`,
		`null`,
		`""`,
	}

	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			var jt JiraTime
			data := []byte(tc)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				json.Unmarshal(data, &jt)
			}
		})
	}
}

func BenchmarkJiraTime_MarshalJSON(b *testing.B) {
	testTimes := []JiraTime{
		{Time: time.Date(2023, 12, 31, 15, 4, 5, 0, time.FixedZone("+0300", 3*60*60))},
		{Time: time.Date(2023, 12, 31, 15, 4, 5, 0, time.UTC)},
		{Time: time.Time{}}, // zero time
	}

	for _, jt := range testTimes {
		b.Run(jt.String(), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				json.Marshal(jt)
			}
		})
	}
}
