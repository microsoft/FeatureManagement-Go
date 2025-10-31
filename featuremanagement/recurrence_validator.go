// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package featuremanagement

import (
	"fmt"
	"time"
)

const (
	valueOutOfRangeErrMsg              = "The value is out of the accepted range."
	unrecognizableValueErrMsg          = "The value is unrecognizable."
	requiredParameterMissingErrMsg     = "Value cannot be undefined or empty."
	startNotMatchedErrMsg              = "Start date is not a valid first occurrence."
	timeWindowDurationOutOfRangeErrMsg = "Time window duration cannot be longer than how frequently it occurs or be longer than 10 years."

	recurrencePatternType    = "Recurrence.Pattern.Type"
	recurrenceInterval       = "Recurrence.Pattern.Interval"
	recurrenceDaysOfWeek     = "Recurrence.Pattern.DaysOfWeek"
	recurrenceFirstDayOfWeek = "Recurrence.Pattern.FirstDayOfWeek"
	recurrenceRangeType      = "Recurrence.Range.Type"
	recurrenceEndDate        = "Recurrence.Range.EndDate"
	numberOfRecurrences      = "Recurrence.Range.NumberOfOccurrences"
)

// buildInvalidParameterError creates an error message for invalid parameters
func buildInvalidParameterError(parameterName, additionalInfo string) error {
	return fmt.Errorf("the %s parameter is not valid. %s", parameterName, additionalInfo)
}

// parseRecurrenceParameter parses RecurrenceParameters into a RecurrenceSpec object
// If the parameter is invalid, an error will be returned
func parseRecurrenceParameter(startTime, endTime *time.Time, recurrenceParams *RecurrenceParameters, timezoneOffset int64) (*RecurrenceSpec, error) {
	if startTime == nil {
		return nil, buildInvalidParameterError("Start", requiredParameterMissingErrMsg)
	}
	if endTime == nil {
		return nil, buildInvalidParameterError("End", requiredParameterMissingErrMsg)
	}
	if !startTime.Before(*endTime) {
		return nil, buildInvalidParameterError("End", valueOutOfRangeErrMsg)
	}

	timeWindowDuration := endTime.Sub(*startTime).Milliseconds()
	if timeWindowDuration > 10*365*OneDayInMilliSeconds {
		return nil, buildInvalidParameterError("End", timeWindowDurationOutOfRangeErrMsg)
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
		return nil, buildInvalidParameterError(recurrencePatternType, requiredParameterMissingErrMsg)
	}

	patternType, ok := ParseRecurrencePatternType(rawPattern.Type)
	if !ok {
		return nil, buildInvalidParameterError(recurrencePatternType, unrecognizableValueErrMsg)
	}

	interval := 1
	if rawPattern.Interval != nil {
		if *rawPattern.Interval <= 0 {
			return nil, buildInvalidParameterError(recurrenceInterval, valueOutOfRangeErrMsg)
		}
		interval = *rawPattern.Interval
	}

	pattern := &RecurrencePattern{
		Type:     patternType,
		Interval: interval,
	}

	timeWindowDuration := endTime.Sub(*startTime).Milliseconds()

	if patternType == Daily {
		if timeWindowDuration > int64(interval)*OneDayInMilliSeconds {
			return nil, buildInvalidParameterError("End", timeWindowDurationOutOfRangeErrMsg)
		}
	} else if patternType == Weekly {
		// Parse FirstDayOfWeek
		firstDayOfWeek := Sunday
		if rawPattern.FirstDayOfWeek != nil {
			day, ok := ParseDayOfWeek(*rawPattern.FirstDayOfWeek)
			if !ok {
				return nil, buildInvalidParameterError(recurrenceFirstDayOfWeek, unrecognizableValueErrMsg)
			}
			firstDayOfWeek = day
		}
		pattern.FirstDayOfWeek = firstDayOfWeek

		// Parse DaysOfWeek
		if len(rawPattern.DaysOfWeek) == 0 {
			return nil, buildInvalidParameterError(recurrenceDaysOfWeek, requiredParameterMissingErrMsg)
		}

		daysMap := make(map[DayOfWeek]bool)
		var daysOfWeek []DayOfWeek
		for _, dayStr := range rawPattern.DaysOfWeek {
			day, ok := ParseDayOfWeek(dayStr)
			if !ok {
				return nil, buildInvalidParameterError(recurrenceDaysOfWeek, unrecognizableValueErrMsg)
			}
			// Deduplicate
			if !daysMap[day] {
				daysMap[day] = true
				daysOfWeek = append(daysOfWeek, day)
			}
		}

		if len(daysOfWeek) == 0 {
			return nil, buildInvalidParameterError(recurrenceDaysOfWeek, requiredParameterMissingErrMsg)
		}

		// Check duration constraint
		if timeWindowDuration > int64(interval)*DaysPerWeek*OneDayInMilliSeconds ||
			!isDurationCompliantWithDaysOfWeek(timeWindowDuration, interval, daysOfWeek, firstDayOfWeek) {
			return nil, buildInvalidParameterError("End", timeWindowDurationOutOfRangeErrMsg)
		}

		pattern.DaysOfWeek = daysOfWeek

		// Check whether "Start" is a valid first occurrence
		alignedStartDay := getDayOfWeek(*startTime, timezoneOffset)
		found := false
		for _, day := range daysOfWeek {
			if day == alignedStartDay {
				found = true
				break
			}
		}
		if !found {
			return nil, buildInvalidParameterError("Start", startNotMatchedErrMsg)
		}
	}

	return pattern, nil
}

// parseRecurrenceRange parses and validates the recurrence range
func parseRecurrenceRange(startTime *time.Time, params *RecurrenceParameters) (*RecurrenceRange, error) {
	rawRange := params.Range

	if rawRange.Type == "" {
		return nil, buildInvalidParameterError(recurrenceRangeType, requiredParameterMissingErrMsg)
	}

	rangeType, ok := ParseRecurrenceRangeType(rawRange.Type)
	if !ok {
		return nil, buildInvalidParameterError(recurrenceRangeType, unrecognizableValueErrMsg)
	}

	recurrenceRange := &RecurrenceRange{
		Type: rangeType,
	}

	if rangeType == EndDate {
		var endDate time.Time
		if rawRange.EndDate != nil {
			parsed, err := parseTime(*rawRange.EndDate)
			if err != nil {
				return nil, buildInvalidParameterError(recurrenceEndDate, unrecognizableValueErrMsg)
			}
			endDate = parsed
			if endDate.Before(*startTime) {
				return nil, buildInvalidParameterError(recurrenceEndDate, valueOutOfRangeErrMsg)
			}
		} else {
			// Maximum date in Go (year 9999)
			endDate = time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC)
		}
		recurrenceRange.EndDate = &endDate
	} else if rangeType == Numbered {
		numberOfOccurrences := int(^uint(0) >> 1) // Max int
		if rawRange.NumberOfOccurrences != nil {
			if *rawRange.NumberOfOccurrences <= 0 {
				return nil, buildInvalidParameterError(numberOfRecurrences, valueOutOfRangeErrMsg)
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
	minGap := int64(DaysPerWeek * OneDayInMilliSeconds)

	for i := 1; i < len(sortedDaysOfWeek); i++ { // skip the first day
		gap := int64(calculateWeeklyDayOffset(sortedDaysOfWeek[i], prev)) * OneDayInMilliSeconds
		if gap < minGap {
			minGap = gap
		}
		prev = sortedDaysOfWeek[i]
	}

	// It may cross weeks. Check the next week if the interval is one week.
	if interval == 1 {
		gap := int64(calculateWeeklyDayOffset(sortedDaysOfWeek[0], prev)) * OneDayInMilliSeconds
		if gap < minGap {
			minGap = gap
		}
	}

	return minGap >= duration
}
