// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package featuremanagement

import "time"

const (
	// DaysPerWeek is the number of days in a week
	DaysPerWeek = 7
	// OneDayInMilliSeconds is the number of milliseconds in one day
	OneDayInMilliSeconds = 24 * 60 * 60 * 1000
)

// DayOfWeek represents a day of the week (0 = Sunday, 6 = Saturday)
type DayOfWeek int

const (
	Sunday DayOfWeek = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

// String returns the string representation of the day of week
func (d DayOfWeek) String() string {
	return [...]string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}[d]
}

// ParseDayOfWeek converts a string to a DayOfWeek
func ParseDayOfWeek(s string) (DayOfWeek, bool) {
	dayMap := map[string]DayOfWeek{
		"Sunday":    Sunday,
		"Monday":    Monday,
		"Tuesday":   Tuesday,
		"Wednesday": Wednesday,
		"Thursday":  Thursday,
		"Friday":    Friday,
		"Saturday":  Saturday,
	}
	day, ok := dayMap[s]
	return day, ok
}

// RecurrencePatternType describes the frequency by which the time window repeats
type RecurrencePatternType int

const (
	// Daily pattern repeats based on the number of days specified by interval between occurrences
	Daily RecurrencePatternType = iota
	// Weekly pattern repeats on the same day or days of the week, based on the number of weeks between each set of occurrences
	Weekly
)

// String returns the string representation of the recurrence pattern type
func (r RecurrencePatternType) String() string {
	return [...]string{"Daily", "Weekly"}[r]
}

// ParseRecurrencePatternType converts a string to a RecurrencePatternType
func ParseRecurrencePatternType(s string) (RecurrencePatternType, bool) {
	typeMap := map[string]RecurrencePatternType{
		"Daily":  Daily,
		"Weekly": Weekly,
	}
	t, ok := typeMap[s]
	return t, ok
}

// RecurrenceRangeType specifies the date range over which the time window repeats
type RecurrenceRangeType int

const (
	// NoEnd recurrence has no end and repeats on all days that fit the corresponding pattern
	NoEnd RecurrenceRangeType = iota
	// EndDate recurrence repeats on all days that fit the pattern until or on the specified end date
	EndDate
	// Numbered recurrence repeats for the specified number of occurrences that match the pattern
	Numbered
)

// String returns the string representation of the recurrence range type
func (r RecurrenceRangeType) String() string {
	return [...]string{"NoEnd", "EndDate", "Numbered"}[r]
}

// ParseRecurrenceRangeType converts a string to a RecurrenceRangeType
func ParseRecurrenceRangeType(s string) (RecurrenceRangeType, bool) {
	typeMap := map[string]RecurrenceRangeType{
		"NoEnd":    NoEnd,
		"EndDate":  EndDate,
		"Numbered": Numbered,
	}
	t, ok := typeMap[s]
	return t, ok
}

// RecurrencePattern describes the frequency by which the time window repeats
type RecurrencePattern struct {
	// Type is the type of the recurrence pattern
	Type RecurrencePatternType
	// Interval is the number of units between occurrences (days or weeks depending on pattern type)
	Interval int
	// DaysOfWeek are the days when the time window occurs (only for Weekly pattern)
	DaysOfWeek []DayOfWeek
	// FirstDayOfWeek is the first day of the week (only for Weekly pattern)
	FirstDayOfWeek DayOfWeek
}

// RecurrenceRange describes the date range over which the time window repeats
type RecurrenceRange struct {
	// Type is the type of the recurrence range
	Type RecurrenceRangeType
	// EndDate is the date to stop applying the recurrence pattern (only for EndDate range)
	EndDate *time.Time
	// NumberOfOccurrences is the number of times to repeat the time window (only for Numbered range)
	NumberOfOccurrences *int
}

// RecurrenceSpec defines the complete recurring time window specification
type RecurrenceSpec struct {
	// StartTime is the start time of the first/base time window
	StartTime time.Time
	// Duration is the duration of each time window in milliseconds
	Duration int64
	// Pattern is the recurrence pattern
	Pattern RecurrencePattern
	// Range is the recurrence range
	Range RecurrenceRange
	// TimezoneOffset is the timezone offset in milliseconds for day-of-week calculations
	TimezoneOffset int64
}

// RecurrenceParameters represents the JSON parameters for recurrence configuration
type RecurrenceParameters struct {
	Pattern RecurrencePatternParams
	Range   RecurrenceRangeParams
}

// RecurrencePatternParams represents the JSON pattern parameters
type RecurrencePatternParams struct {
	Type           string
	Interval       *int
	DaysOfWeek     []string
	FirstDayOfWeek *string
}

// RecurrenceRangeParams represents the JSON range parameters
type RecurrenceRangeParams struct {
	Type                string
	EndDate             *string
	NumberOfOccurrences *int
}

// recurrenceState tracks the state of recurrence matching
type recurrenceState struct {
	PreviousOccurrence  time.Time
	NumberOfOccurrences int
}
