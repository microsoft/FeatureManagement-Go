// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package featuremanagement

import (
	"testing"
	"time"
)

// Helper function to create a pointer to an int
func intPtr(i int) *int {
	return &i
}

// Helper function to create a pointer to a string
func stringPtr(s string) *string {
	return &s
}

func TestRecurrenceValidator_RequiredParameters(t *testing.T) {
	// Test missing Start parameter
	recurrence1 := &RecurrenceParameters{
		Pattern: RecurrencePatternParams{Type: "Daily"},
		Range:   RecurrenceRangeParams{Type: "NoEnd"},
	}
	endTime := time.Date(2023, 9, 1, 0, 0, 1, 0, time.UTC)
	_, err := parseRecurrenceParameter(nil, &endTime, recurrence1, 0)
	if err == nil || err.Error() != "the Start parameter is required for recurrence configuration" {
		t.Errorf("Expected Start required error, got: %v", err)
	}

	// Test missing End parameter
	startTime := time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC)
	_, err = parseRecurrenceParameter(&startTime, nil, recurrence1, 0)
	if err == nil || err.Error() != "the End parameter is required for recurrence configuration" {
		t.Errorf("Expected End required error, got: %v", err)
	}

	// Test missing Pattern
	recurrence2 := &RecurrenceParameters{
		Range: RecurrenceRangeParams{Type: "NoEnd"},
	}
	startTime2 := time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC)
	endTime2 := time.Date(2023, 9, 1, 0, 0, 1, 0, time.UTC)
	_, err = parseRecurrenceParameter(&startTime2, &endTime2, recurrence2, 0)
	if err == nil {
		t.Error("Expected Pattern.Type required error")
	}

	// Test missing DaysOfWeek for Weekly pattern
	recurrence3 := &RecurrenceParameters{
		Pattern: RecurrencePatternParams{Type: "Weekly"},
		Range:   RecurrenceRangeParams{Type: "NoEnd"},
	}
	_, err = parseRecurrenceParameter(&startTime2, &endTime2, recurrence3, 0)
	if err == nil {
		t.Error("Expected DaysOfWeek required error")
	}
}

func TestRecurrenceValidator_InvalidValues(t *testing.T) {
	// Test invalid Interval
	recurrence1 := &RecurrenceParameters{
		Pattern: RecurrencePatternParams{
			Type:     "Daily",
			Interval: intPtr(0),
		},
		Range: RecurrenceRangeParams{Type: "NoEnd"},
	}
	startTime := time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 9, 1, 0, 0, 1, 0, time.UTC)
	_, err := parseRecurrenceParameter(&startTime, &endTime, recurrence1, 0)
	if err == nil {
		t.Error("Expected Interval out of range error")
	}

	// Test invalid NumberOfOccurrences
	recurrence2 := &RecurrenceParameters{
		Pattern: RecurrencePatternParams{Type: "Daily"},
		Range: RecurrenceRangeParams{
			Type:                "Numbered",
			NumberOfOccurrences: intPtr(0),
		},
	}
	_, err = parseRecurrenceParameter(&startTime, &endTime, recurrence2, 0)
	if err == nil {
		t.Error("Expected NumberOfOccurrences out of range error")
	}

	// Test invalid DayOfWeek
	recurrence3 := &RecurrenceParameters{
		Pattern: RecurrencePatternParams{
			Type:       "Weekly",
			DaysOfWeek: []string{"Monday", "Tue"},
		},
		Range: RecurrenceRangeParams{Type: "NoEnd"},
	}
	startTime3 := time.Date(2023, 9, 4, 0, 0, 0, 0, time.UTC) // Monday
	endTime3 := time.Date(2023, 9, 4, 0, 0, 1, 0, time.UTC)
	_, err = parseRecurrenceParameter(&startTime3, &endTime3, recurrence3, 0)
	if err == nil {
		t.Error("Expected DaysOfWeek unrecognizable error")
	}

	// Test invalid FirstDayOfWeek
	recurrence4 := &RecurrenceParameters{
		Pattern: RecurrencePatternParams{
			Type:           "Weekly",
			DaysOfWeek:     []string{"Monday"},
			FirstDayOfWeek: stringPtr("Mon"),
		},
		Range: RecurrenceRangeParams{Type: "NoEnd"},
	}
	_, err = parseRecurrenceParameter(&startTime3, &endTime3, recurrence4, 0)
	if err == nil {
		t.Error("Expected FirstDayOfWeek unrecognizable error")
	}

	// Test invalid EndDate
	recurrence5 := &RecurrenceParameters{
		Pattern: RecurrencePatternParams{Type: "Daily"},
		Range: RecurrenceRangeParams{
			Type:    "EndDate",
			EndDate: stringPtr("AppConfig"),
		},
	}
	_, err = parseRecurrenceParameter(&startTime, &endTime, recurrence5, 0)
	if err == nil {
		t.Error("Expected EndDate unrecognizable error")
	}
}

func TestRecurrenceValidator_TimeWindowDuration(t *testing.T) {
	// Test End before Start
	recurrence := &RecurrenceParameters{
		Pattern: RecurrencePatternParams{Type: "Daily"},
		Range:   RecurrenceRangeParams{Type: "NoEnd"},
	}
	startTime := time.Date(2023, 9, 2, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC)
	_, err := parseRecurrenceParameter(&startTime, &endTime, recurrence, 0)
	if err == nil {
		t.Error("Expected End out of range error")
	}

	// Test duration too long for Daily pattern
	startTime2 := time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC)
	endTime2 := time.Date(2024, 12, 12, 0, 0, 0, 0, time.UTC)
	_, err = parseRecurrenceParameter(&startTime2, &endTime2, recurrence, 0)
	if err == nil {
		t.Error("Expected time window duration out of range error")
	}

	// Test valid Weekly pattern
	recurrence2 := &RecurrenceParameters{
		Pattern: RecurrencePatternParams{
			Type:       "Weekly",
			DaysOfWeek: []string{"Monday"},
		},
		Range: RecurrenceRangeParams{Type: "NoEnd"},
	}
	startTime3 := time.Date(2024, 12, 9, 0, 0, 0, 0, time.UTC) // Monday
	endTime3 := time.Date(2024, 12, 16, 0, 0, 0, 0, time.UTC)
	_, err = parseRecurrenceParameter(&startTime3, &endTime3, recurrence2, 0)
	if err != nil {
		t.Errorf("Expected no error for valid weekly pattern, got: %v", err)
	}
}

func TestRecurrenceEvaluator_DailyRecurrence(t *testing.T) {
	// Test basic daily recurrence
	startTime := time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC)
	spec := &RecurrenceSpec{
		StartTime: startTime,
		Duration:  1000,
		Pattern: RecurrencePattern{
			Type:     Daily,
			Interval: 1,
		},
		Range: RecurrenceRange{
			Type: NoEnd,
		},
		TimezoneOffset: 0,
	}

	testTime := time.Date(2023, 9, 2, 0, 0, 0, 0, time.UTC)
	if !matchRecurrence(testTime, spec) {
		t.Error("Expected daily recurrence to match")
	}

	// Test with interval of 2
	spec.Pattern.Interval = 2
	testTime2 := time.Date(2023, 9, 2, 0, 0, 0, 0, time.UTC)
	if matchRecurrence(testTime2, spec) {
		t.Error("Expected no match with interval 2")
	}

	testTime3 := time.Date(2023, 9, 3, 0, 0, 0, 0, time.UTC)
	if !matchRecurrence(testTime3, spec) {
		t.Error("Expected match on day 3 with interval 2")
	}

	// Test with NumberOfOccurrences
	spec.Pattern.Interval = 1
	occurrences := 2
	spec.Range = RecurrenceRange{
		Type:                Numbered,
		NumberOfOccurrences: &occurrences,
	}

	testTime4 := time.Date(2023, 9, 2, 0, 0, 0, 0, time.UTC)
	if !matchRecurrence(testTime4, spec) {
		t.Error("Expected match within occurrence limit")
	}

	testTime5 := time.Date(2023, 9, 3, 0, 0, 0, 0, time.UTC)
	if matchRecurrence(testTime5, spec) {
		t.Error("Expected no match beyond occurrence limit")
	}
}

func TestRecurrenceEvaluator_WeeklyRecurrence(t *testing.T) {
	// Test weekly recurrence on Monday and Friday
	startTime := time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC) // Friday
	spec := &RecurrenceSpec{
		StartTime: startTime,
		Duration:  1000,
		Pattern: RecurrencePattern{
			Type:           Weekly,
			Interval:       1,
			DaysOfWeek:     []DayOfWeek{Monday, Friday},
			FirstDayOfWeek: Sunday,
		},
		Range: RecurrenceRange{
			Type: NoEnd,
		},
		TimezoneOffset: 0,
	}

	// Test Monday
	testTime1 := time.Date(2023, 9, 4, 0, 0, 0, 0, time.UTC)
	if !matchRecurrence(testTime1, spec) {
		t.Error("Expected match on Monday")
	}

	// Test Friday
	testTime2 := time.Date(2023, 9, 8, 0, 0, 0, 0, time.UTC)
	if !matchRecurrence(testTime2, spec) {
		t.Error("Expected match on Friday")
	}

	// Test with interval of 2
	spec.Pattern.Interval = 2
	testTime3 := time.Date(2023, 9, 8, 0, 0, 0, 0, time.UTC)
	if matchRecurrence(testTime3, spec) {
		t.Error("Expected no match with interval 2 on week 2")
	}

	testTime4 := time.Date(2023, 9, 15, 0, 0, 0, 0, time.UTC)
	if !matchRecurrence(testTime4, spec) {
		t.Error("Expected match on week 3 with interval 2")
	}

	// Test with Numbered occurrences
	spec.Pattern.Interval = 1
	occurrences := 1
	spec.Range = RecurrenceRange{
		Type:                Numbered,
		NumberOfOccurrences: &occurrences,
	}

	testTime5 := time.Date(2023, 9, 4, 0, 0, 0, 0, time.UTC)
	if matchRecurrence(testTime5, spec) {
		t.Error("Expected no match - first occurrence is on start day (Friday)")
	}
}

func TestRecurrenceEvaluator_WeeklyRecurrenceWithFirstDayOfWeek(t *testing.T) {
	// Test with FirstDayOfWeek = Monday
	startTime := time.Date(2023, 9, 3, 0, 0, 0, 0, time.UTC) // Sunday
	spec := &RecurrenceSpec{
		StartTime: startTime,
		Duration:  1000,
		Pattern: RecurrencePattern{
			Type:           Weekly,
			Interval:       2,
			DaysOfWeek:     []DayOfWeek{Sunday, Monday},
			FirstDayOfWeek: Monday,
		},
		Range: RecurrenceRange{
			Type: NoEnd,
		},
		TimezoneOffset: 0,
	}

	// Test Monday (should not match - different week in interval)
	testTime1 := time.Date(2023, 9, 4, 0, 0, 0, 0, time.UTC)
	if matchRecurrence(testTime1, spec) {
		t.Error("Expected no match on Monday in wrong week")
	}

	// Test Sunday in the correct interval
	testTime2 := time.Date(2023, 9, 17, 0, 0, 0, 0, time.UTC)
	if !matchRecurrence(testTime2, spec) {
		t.Error("Expected match on Sunday in correct week")
	}
}

func TestRecurrenceEvaluator_TimezoneHandling(t *testing.T) {
	// Test with timezone offset (8 hours = 8 * 60 * 60 * 1000 ms)
	timezoneOffset := int64(8 * 60 * 60 * 1000)
	// Start on Friday Sept 1, 2023 00:00:00 UTC (which is Friday in UTC+8 also)
	startTime := time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC)
	spec := &RecurrenceSpec{
		StartTime: startTime,
		Duration:  1000,
		Pattern: RecurrencePattern{
			Type:           Weekly,
			Interval:       1,
			DaysOfWeek:     []DayOfWeek{Friday},
			FirstDayOfWeek: Sunday,
		},
		Range: RecurrenceRange{
			Type: NoEnd,
		},
		TimezoneOffset: timezoneOffset,
	}

	// Test on the start day (Friday)
	testTime := time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC)
	if !matchRecurrence(testTime, spec) {
		t.Error("Expected match on start day (Friday)")
	}

	// Test next Friday (Sept 8, 2023 00:00:00 UTC = Sept 8, 2023 08:00 UTC+8, still Friday)
	testTime2 := time.Date(2023, 9, 8, 0, 0, 0, 0, time.UTC)
	if !matchRecurrence(testTime2, spec) {
		t.Error("Expected match on next Friday")
	}

	// Test  a non-Friday (Monday Sept 4)
	testTime3 := time.Date(2023, 9, 4, 0, 0, 0, 0, time.UTC)
	if matchRecurrence(testTime3, spec) {
		t.Error("Expected no match on Monday")
	}
}
