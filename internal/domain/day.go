package domain

import "time"

// Day represents a day with word count
type Day struct {
	Date      time.Time
	WordCount int
}

// DateString returns date in YYYYMMDD format
func (d Day) DateString() string {
	return d.Date.Format("20060102")
}

// DisplayString returns user-friendly date string
func (d Day) DisplayString() string {
	now := time.Now()
	date := d.Date

	// Check if today
	if date.Year() == now.Year() && date.Month() == now.Month() && date.Day() == now.Day() {
		return "Сегодня"
	}

	// Check if yesterday
	yesterday := now.AddDate(0, 0, -1)
	if date.Year() == yesterday.Year() && date.Month() == yesterday.Month() && date.Day() == yesterday.Day() {
		return "Вчера"
	}

	// Return formatted date
	months := []string{
		"", "янв", "фев", "мар", "апр", "мая", "июн",
		"июл", "авг", "сен", "окт", "ноя", "дек",
	}

	return date.Format("2 ") + months[date.Month()] + date.Format(" 2006")
}

