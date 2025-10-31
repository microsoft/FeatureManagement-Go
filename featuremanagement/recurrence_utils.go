// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package featuremanagement

import "time"

// calculateWeeklyDayOffset calculates the offset in days between two given days of the week
// Returns the number of days to be added to day2 to reach day1
func calculateWeeklyDayOffset(day1, day2 DayOfWeek) int {
	return (int(day1) - int(day2) + DaysPerWeek) % DaysPerWeek
}

// sortDaysOfWeek sorts a collection of days of week based on their offsets from a specified first day of week
// The first day of week will be the first element in the sorted result
func sortDaysOfWeek(daysOfWeek []DayOfWeek, firstDayOfWeek DayOfWeek) []DayOfWeek {
	// Create a copy to avoid modifying the original slice
	sorted := make([]DayOfWeek, len(daysOfWeek))
	copy(sorted, daysOfWeek)

	// Sort by offset from first day of week
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			offset1 := calculateWeeklyDayOffset(sorted[i], firstDayOfWeek)
			offset2 := calculateWeeklyDayOffset(sorted[j], firstDayOfWeek)
			if offset1 > offset2 {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

// getDayOfWeek gets the day of week of a given date based on the timezone offset
// timezoneOffsetInMs is the timezone offset in milliseconds
// Returns the day of week (0 for Sunday, 1 for Monday, ..., 6 for Saturday)
func getDayOfWeek(date time.Time, timezoneOffsetInMs int64) DayOfWeek {
	// Apply timezone offset
	alignedDate := date.Add(time.Duration(timezoneOffsetInMs) * time.Millisecond)
	// Get the weekday (Sunday = 0, Monday = 1, etc.)
	return DayOfWeek(alignedDate.Weekday())
}

// addDays adds a specified number of days to a given date
func addDays(date time.Time, days int) time.Time {
	return date.Add(time.Duration(days) * 24 * time.Hour)
}
