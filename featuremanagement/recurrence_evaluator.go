// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package featuremanagement

import (
	"math"
	"time"
)

// matchRecurrence checks if a provided datetime is within any recurring time window
// specified by the recurrence information
func matchRecurrence(currentTime time.Time, spec *RecurrenceSpec) bool {
	state := findPreviousRecurrence(currentTime, spec)
	if state != nil {
		windowEnd := state.PreviousOccurrence.Add(time.Duration(spec.Duration) * time.Millisecond)
		return currentTime.Before(windowEnd)
	}
	return false
}

// findPreviousRecurrence finds the closest previous recurrence occurrence before the given time
// according to the recurrence information
func findPreviousRecurrence(currentTime time.Time, spec *RecurrenceSpec) *recurrenceState {
	if currentTime.Before(spec.StartTime) {
		return nil
	}

	var state *recurrenceState

	switch spec.Pattern.Type {
	case Daily:
		state = findPreviousDailyRecurrence(currentTime, spec)
	case Weekly:
		state = findPreviousWeeklyRecurrence(currentTime, spec)
	default:
		return nil
	}

	if state == nil {
		return nil
	}

	// Check range constraints
	switch spec.Range.Type {
	case EndDate:
		if spec.Range.EndDate != nil && state.PreviousOccurrence.After(*spec.Range.EndDate) {
			return nil
		}
	case Numbered:
		if spec.Range.NumberOfOccurrences != nil && state.NumberOfOccurrences > *spec.Range.NumberOfOccurrences {
			return nil
		}
	}

	return state
}

// findPreviousDailyRecurrence finds the previous occurrence for daily recurrence pattern
func findPreviousDailyRecurrence(currentTime time.Time, spec *RecurrenceSpec) *recurrenceState {
	timeGap := currentTime.Sub(spec.StartTime).Milliseconds()
	numberOfIntervals := int(math.Floor(float64(timeGap) / float64(spec.Pattern.Interval*oneDayInMilliSeconds)))

	return &recurrenceState{
		PreviousOccurrence:  addDays(spec.StartTime, numberOfIntervals*spec.Pattern.Interval),
		NumberOfOccurrences: numberOfIntervals + 1,
	}
}

// findPreviousWeeklyRecurrence finds the previous occurrence for weekly recurrence pattern
func findPreviousWeeklyRecurrence(currentTime time.Time, spec *RecurrenceSpec) *recurrenceState {
	/*
	 * Algorithm:
	 * 1. first find day 0 (d0), it's the day representing the start day on the week of `Start`.
	 * 2. find start day of the most recent occurring week d0 + floor((time - d0) / (interval * 7)) * (interval * 7)
	 * 3. if that's over 7 days ago, then previous occurrence is the day with the max offset of the last occurring week
	 * 4. if gotten this far, then the current week is the most recent occurring week:
	 *     i. if time > day with min offset, then previous occurrence is the day with max offset less than current
	 *    ii. if time < day with min offset, then previous occurrence is the day with the max offset of previous occurring week
	 */

	startDay := getDayOfWeek(spec.StartTime, spec.TimezoneOffset)
	sortedDaysOfWeek := sortDaysOfWeek(spec.Pattern.DaysOfWeek, spec.Pattern.FirstDayOfWeek)

	// Calculate the first day of the week containing the start time
	firstDayOfStartWeek := addDays(spec.StartTime, -calculateWeeklyDayOffset(startDay, spec.Pattern.FirstDayOfWeek))

	timeGap := currentTime.Sub(firstDayOfStartWeek).Milliseconds()

	// Number of intervals before the most recent occurring week
	numberOfIntervals := int(math.Floor(float64(timeGap) / float64(spec.Pattern.Interval*daysPerWeek*oneDayInMilliSeconds)))

	// Number of occurrences before the most recent occurring week (can be negative)
	numberOfOccurrences := numberOfIntervals*len(sortedDaysOfWeek) - findDayIndex(sortedDaysOfWeek, startDay)

	// First day of the latest occurring week
	firstDayOfLatestOccurringWeek := addDays(firstDayOfStartWeek, numberOfIntervals*spec.Pattern.Interval*daysPerWeek)

	// Check if current time is beyond the last occurring week
	weekBoundary := addDays(firstDayOfLatestOccurringWeek, daysPerWeek)
	if currentTime.After(weekBoundary) {
		numberOfOccurrences += len(sortedDaysOfWeek)
		// Day with max offset in the last occurring week
		lastDay := sortedDaysOfWeek[len(sortedDaysOfWeek)-1]
		previousOccurrence := addDays(firstDayOfLatestOccurringWeek, calculateWeeklyDayOffset(lastDay, spec.Pattern.FirstDayOfWeek))
		return &recurrenceState{
			PreviousOccurrence:  previousOccurrence,
			NumberOfOccurrences: numberOfOccurrences,
		}
	}

	// Current week is the most recent occurring week
	dayWithMinOffset := addDays(firstDayOfLatestOccurringWeek, calculateWeeklyDayOffset(sortedDaysOfWeek[0], spec.Pattern.FirstDayOfWeek))
	if dayWithMinOffset.Before(spec.StartTime) {
		numberOfOccurrences = 0
		dayWithMinOffset = spec.StartTime
	}

	var previousOccurrence time.Time
	if !currentTime.Before(dayWithMinOffset) {
		// Previous occurrence is the day with max offset less than current
		previousOccurrence = dayWithMinOffset
		numberOfOccurrences++

		dayWithMinOffsetIndex := findDayIndex(sortedDaysOfWeek, getDayOfWeek(dayWithMinOffset, spec.TimezoneOffset))
		for i := dayWithMinOffsetIndex + 1; i < len(sortedDaysOfWeek); i++ {
			day := addDays(firstDayOfLatestOccurringWeek, calculateWeeklyDayOffset(sortedDaysOfWeek[i], spec.Pattern.FirstDayOfWeek))
			if !currentTime.Before(day) {
				previousOccurrence = day
				numberOfOccurrences++
			} else {
				break
			}
		}
	} else {
		// Previous occurrence is the day with the max offset of previous occurring week
		firstDayOfPreviousOccurringWeek := addDays(firstDayOfLatestOccurringWeek, -spec.Pattern.Interval*daysPerWeek)
		lastDay := sortedDaysOfWeek[len(sortedDaysOfWeek)-1]
		previousOccurrence = addDays(firstDayOfPreviousOccurringWeek, calculateWeeklyDayOffset(lastDay, spec.Pattern.FirstDayOfWeek))
	}

	return &recurrenceState{
		PreviousOccurrence:  previousOccurrence,
		NumberOfOccurrences: numberOfOccurrences,
	}
}

// findDayIndex finds the index of a day in the sorted days array
func findDayIndex(days []DayOfWeek, day DayOfWeek) int {
	for i, d := range days {
		if d == day {
			return i
		}
	}
	return 0
}
