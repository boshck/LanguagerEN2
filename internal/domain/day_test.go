package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDay_DateString(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		expected string
	}{
		{
			name:     "date 2024-12-12",
			date:     time.Date(2024, 12, 12, 10, 0, 0, 0, time.UTC),
			expected: "20241212",
		},
		{
			name:     "date 2024-01-01",
			date:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: "20240101",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			day := Day{Date: tt.date}
			result := day.DateString()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDay_DisplayString(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	twoDaysAgo := now.AddDate(0, 0, -2)

	tests := []struct {
		name     string
		date     time.Time
		expected string
	}{
		{
			name:     "today",
			date:     now,
			expected: "Сегодня",
		},
		{
			name:     "yesterday",
			date:     yesterday,
			expected: "Вчера",
		},
		{
			name:     "two days ago",
			date:     twoDaysAgo,
			expected: twoDaysAgo.Format("2 ") + getMonthName(twoDaysAgo.Month()) + twoDaysAgo.Format(" 2006"),
		},
		{
			name:     "specific date",
			date:     time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
			expected: "15 июн 2024",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			day := Day{Date: tt.date}
			result := day.DisplayString()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to get month name in Russian
func getMonthName(m time.Month) string {
	months := []string{
		"", "янв", "фев", "мар", "апр", "мая", "июн",
		"июл", "авг", "сен", "окт", "ноя", "дек",
	}
	return months[m]
}

