// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package featuremanagement

import (
	"fmt"
	"slices"
	"time"
)

const (
	valueOutOfRangeErrMsg              = "The value is out of the accepted range"
	unrecognizableValueErrMsg          = "The value is unrecognizable"
	requiredParameterMissingErrMsg     = "Value cannot be undefined or empty"
	startNotMatchedErrMsg              = "Start date is not a valid first occurrence"
	timeWindowDurationOutOfRangeErrMsg = "Time window duration cannot be longer than how frequently it occurs or be longer than 10 years"

	recurrencePatternType    = "Recurrence.Pattern.Type"
	recurrenceInterval       = "Recurrence.Pattern.Interval"
	recurrenceDaysOfWeek     = "Recurrence.Pattern.DaysOfWeek"
	recurrenceFirstDayOfWeek = "Recurrence.Pattern.FirstDayOfWeek"
	recurrenceRangeType      = "Recurrence.Range.Type"
	recurrenceEndDate        = "Recurrence.Range.EndDate"
	numberOfRecurrences      = "Recurrence.Range.NumberOfOccurrences"
)

// parseRecurrenceParameter parses RecurrenceParameters into a RecurrenceSpec object
// If the parameter is invalid, an error will be returned
func parseRecurrenceParameter(startTime, endTime *time.Time, recurrenceParams *RecurrenceParameters, timezoneOffset int64) (*RecurrenceSpec, error) {
	if startTime == nil {
		return nil, fmt.Errorf("the Start parameter is required for recurrence configuration")
	}
	if endTime == nil {
		return nil, fmt.Errorf("the End parameter is required for recurrence configuration")
	}
	if !startTime.Before(*endTime) {
		return nil, fmt.Errorf("the Start parameter must be before the End parameter for recurrence configuration")
	}

	timeWindowDuration := endTime.Sub(*startTime).Milliseconds()
	if timeWindowDuration > 10*365*oneDayInMilliSeconds {
		return nil, fmt.Errorf("the End parameter is not valid. %s", timeWindowDurationOutOfRangeErrMsg)
	}

	pattern, err := parseRecurrencePattern(startTime, endTime, recurrenceParams, timezoneOffset)
	if err != nil {
		return nil, err
	}

	recurrenceRange, err := parseRecurrenceRange(startTime, recurrenceParams)
	if err != nil {
		return nil, err
	}

	return &RecurrenceSpec{
		StartTime:      *startTime,
		Duration:       timeWindowDuration,
		Pattern:        *pattern,
		Range:          *recurrenceRange,
		TimezoneOffset: timezoneOffset,
	}, nil
}

// parseRecurrencePattern parses and validates the recurrence pattern
func parseRecurrencePattern(startTime, endTime *time.Time, params *RecurrenceParameters, timezoneOffset int64) (*RecurrencePattern, error) {
	rawPattern := params.Pattern

	if rawPattern.Type == "" {
		return nil, fmt.Errorf("the %s parameter is not valid. %s", recurrencePatternType, requiredParameterMissingErrMsg)
	}

	patternType, ok := ParseRecurrencePatternType(rawPattern.Type)
	if !ok {
		return nil, fmt.Errorf("the %s parameter is not valid. %s", recurrencePatternType, unrecognizableValueErrMsg)
	}

	interval := 1
	if rawPattern.Interval != nil {
		if *rawPattern.Interval <= 0 {
			return nil, fmt.Errorf("the %s parameter is not valid. %s", recurrenceInterval, valueOutOfRangeErrMsg)
		}
		interval = *rawPattern.Interval
	}

	pattern := &RecurrencePattern{
		Type:     patternType,
		Interval: interval,
	}

	timeWindowDuration := endTime.Sub(*startTime).Milliseconds()

	if patternType == Daily {
		if timeWindowDuration > int64(interval)*oneDayInMilliSeconds {
			return nil, fmt.Errorf("the End parameter is not valid. %s", timeWindowDurationOutOfRangeErrMsg)
		}
	} else if patternType == Weekly {
		// Parse FirstDayOfWeek
		firstDayOfWeek := Sunday
		if rawPattern.FirstDayOfWeek != nil {
			day, ok := parseDayOfWeek(*rawPattern.FirstDayOfWeek)
			if !ok {
				return nil, fmt.Errorf("the %s parameter is not valid. %s", recurrenceFirstDayOfWeek, unrecognizableValueErrMsg)
			}
			firstDayOfWeek = day
		}
		pattern.FirstDayOfWeek = firstDayOfWeek

		// Parse DaysOfWeek
		if len(rawPattern.DaysOfWeek) == 0 {
			return nil, fmt.Errorf("the %s parameter is not valid. %s", recurrenceDaysOfWeek, requiredParameterMissingErrMsg)
		}

		daysMap := make(map[DayOfWeek]bool)
		var daysOfWeek []DayOfWeek
		for _, dayStr := range rawPattern.DaysOfWeek {
			day, ok := parseDayOfWeek(dayStr)
			if !ok {
				return nil, fmt.Errorf("the %s parameter is not valid. %s", recurrenceDaysOfWeek, unrecognizableValueErrMsg)
			}
			// Deduplicate
			if !daysMap[day] {
				daysMap[day] = true
				daysOfWeek = append(daysOfWeek, day)
			}
		}

		if len(daysOfWeek) == 0 {
			return nil, fmt.Errorf("the %s parameter is not valid. %s", recurrenceDaysOfWeek, requiredParameterMissingErrMsg)
		}

		// Check duration constraint
		if timeWindowDuration > int64(interval)*daysPerWeek*oneDayInMilliSeconds ||
			!isDurationCompliantWithDaysOfWeek(timeWindowDuration, interval, daysOfWeek, firstDayOfWeek) {
			return nil, fmt.Errorf("the End parameter is not valid. %s", timeWindowDurationOutOfRangeErrMsg)
		}

		pattern.DaysOfWeek = daysOfWeek

		// Check whether "Start" is a valid first occurrence
		alignedStartDay := getDayOfWeek(*startTime, timezoneOffset)
		found := slices.Contains(daysOfWeek, alignedStartDay)
		if !found {
			return nil, fmt.Errorf("the Start parameter is not valid. %s", startNotMatchedErrMsg)
		}
	}

	return pattern, nil
}

// parseRecurrenceRange parses and validates the recurrence range
func parseRecurrenceRange(startTime *time.Time, params *RecurrenceParameters) (*RecurrenceRange, error) {
	rawRange := params.Range

	if rawRange.Type == "" {
		return nil, fmt.Errorf("the %s parameter is not valid. %s", recurrenceRangeType, requiredParameterMissingErrMsg)
	}

	rangeType, ok := ParseRecurrenceRangeType(rawRange.Type)
	if !ok {
		return nil, fmt.Errorf("the %s parameter is not valid. %s", recurrenceRangeType, unrecognizableValueErrMsg)
	}

	recurrenceRange := &RecurrenceRange{
		Type: rangeType,
	}

	switch rangeType {
	case EndDate:
		var endDate time.Time
		if rawRange.EndDate != nil {
			parsed, err := parseTime(*rawRange.EndDate)
			if err != nil {
				return nil, fmt.Errorf("the %s parameter is not valid. %s", recurrenceEndDate, unrecognizableValueErrMsg)
			}
			endDate = parsed
			if endDate.Before(*startTime) {
				return nil, fmt.Errorf("the %s parameter is not valid. %s", recurrenceEndDate, valueOutOfRangeErrMsg)
			}
		} else {
			// Maximum date in Go (year 9999)
			endDate = time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC)
		}
		recurrenceRange.EndDate = &endDate
	case Numbered:
		numberOfOccurrences := int(^uint(0) >> 1) // Max int
		if rawRange.NumberOfOccurrences != nil {
			if *rawRange.NumberOfOccurrences <= 0 {
				return nil, fmt.Errorf("the %s parameter is not valid. %s", numberOfRecurrences, valueOutOfRangeErrMsg)
			}
			numberOfOccurrences = *rawRange.NumberOfOccurrences
		}
		recurrenceRange.NumberOfOccurrences = &numberOfOccurrences
	}

	return recurrenceRange, nil
}

// isDurationCompliantWithDaysOfWeek checks if the duration is compliant with the days of week pattern
func isDurationCompliantWithDaysOfWeek(duration int64, interval int, daysOfWeek []DayOfWeek, firstDayOfWeek DayOfWeek) bool {
	if len(daysOfWeek) == 1 {
		return true
	}

	sortedDaysOfWeek := sortDaysOfWeek(daysOfWeek, firstDayOfWeek)
	prev := sortedDaysOfWeek[0] // the closest occurrence day to the first day of week
	minGap := int64(daysPerWeek * oneDayInMilliSeconds)

	for i := 1; i < len(sortedDaysOfWeek); i++ { // skip the first day
		gap := int64(calculateWeeklyDayOffset(sortedDaysOfWeek[i], prev)) * oneDayInMilliSeconds
		if gap < minGap {
			minGap = gap
		}
		prev = sortedDaysOfWeek[i]
	}

	// It may cross weeks. Check the next week if the interval is one week.
	if interval == 1 {
		gap := int64(calculateWeeklyDayOffset(sortedDaysOfWeek[0], prev)) * oneDayInMilliSeconds
		if gap < minGap {
			minGap = gap
		}
	}

	return minGap >= duration
}
