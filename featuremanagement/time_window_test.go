// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package featuremanagement

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTimeWindowFilterIntegration(t *testing.T) {
	// Define test feature flags
	jsonData := `{
        "feature_flags": [
            {
                "id": "PastTimeWindow",
                "description": "A feature flag using a time window filter, that is active from 2023-06-29 07:00:00 to 2023-08-30 07:00:00.",
                "enabled": true,
                "conditions": {
                    "client_filters": [
                        {
                            "name": "Microsoft.TimeWindow",
                            "parameters": {
                                "Start": "Thu, 29 Jun 2023 07:00:00 GMT",
                                "End": "Wed, 30 Aug 2023 07:00:00 GMT"
                            }
                        }
                    ]
                }
            },
            {
                "id": "FutureTimeWindow",
                "description": "A feature flag using a time window filter, that is active from 3023-06-27 06:00:00 to 3023-06-28 06:05:00.",
                "enabled": true,
                "conditions": {
                    "client_filters": [
                        {
                            "name": "Microsoft.TimeWindow",
                            "parameters": {
                                "Start": "Fri, 27 Jun 3023 06:00:00 GMT",
                                "End": "Sat, 28 Jun 3023 06:05:00 GMT"
                            }
                        }
                    ]
                }
            },
            {
                "id": "PresentTimeWindow",
                "description": "A feature flag using a time window filter within current time.",
                "enabled": true,
                "conditions": {
                    "client_filters": [
                        {
                            "name": "Microsoft.TimeWindow",
                            "parameters": {
                                "Start": "Thu, 29 Jun 2023 07:00:00 GMT",
                                "End": "Sat, 28 Jun 3023 06:05:00 GMT"
                            }
                        }
                    ]
                }
            }
        ]
    }`

	// Parse flags
	var featureManagement struct {
		FeatureFlags []FeatureFlag `json:"feature_flags"`
	}
	if err := json.Unmarshal([]byte(jsonData), &featureManagement); err != nil {
		t.Fatalf("Failed to unmarshal test data: %v", err)
	}

	// Create mock provider
	provider := &mockFeatureFlagProvider{featureFlags: featureManagement.FeatureFlags}

	// Create feature manager
	manager, err := NewFeatureManager(provider, nil)
	if err != nil {
		t.Fatalf("Failed to create feature manager: %v", err)
	}

	// Test cases
	tests := []struct {
		name         string
		featureID    string
		mockedTime   time.Time
		expectResult bool
	}{
		{
			name:         "Past time window should return false",
			featureID:    "PastTimeWindow",
			expectResult: false,
		},
		{
			name:         "Future time window should return false",
			featureID:    "FutureTimeWindow",
			expectResult: false,
		},
		{
			name:         "Present time window should return true",
			featureID:    "PresentTimeWindow",
			expectResult: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Evaluate the feature flag
			result, err := manager.IsEnabled(tc.featureID)
			if err != nil {
				t.Fatalf("Failed to evaluate feature: %v", err)
			}

			if result != tc.expectResult {
				t.Errorf("Expected result %v but got %v", tc.expectResult, result)
			}
		})
	}
}

func TestTimeWindowFilterWithRecurrence_DailyPattern(t *testing.T) {
	// Define test feature flag with daily recurrence
	jsonData := `{
        "feature_flags": [
            {
                "id": "DailyRecurrence",
                "description": "A feature flag with daily recurrence from 9 AM to 5 PM",
                "enabled": true,
                "conditions": {
                    "client_filters": [
                        {
                            "name": "Microsoft.TimeWindow",
                            "parameters": {
                                "Start": "2023-09-01T09:00:00Z",
                                "End": "2023-09-01T17:00:00Z",
                                "Recurrence": {
                                    "Pattern": {
                                        "Type": "Daily",
                                        "Interval": 1
                                    },
                                    "Range": {
                                        "Type": "NoEnd"
                                    }
                                }
                            }
                        }
                    ]
                }
            },
            {
                "id": "DailyRecurrenceWithLimit",
                "description": "Daily recurrence for 3 days only",
                "enabled": true,
                "conditions": {
                    "client_filters": [
                        {
                            "name": "Microsoft.TimeWindow",
                            "parameters": {
                                "Start": "2023-09-01T09:00:00Z",
                                "End": "2023-09-01T17:00:00Z",
                                "Recurrence": {
                                    "Pattern": {
                                        "Type": "Daily",
                                        "Interval": 1
                                    },
                                    "Range": {
                                        "Type": "Numbered",
                                        "NumberOfOccurrences": 3
                                    }
                                }
                            }
                        }
                    ]
                }
            }
        ]
    }`

	var featureManagement struct {
		FeatureFlags []FeatureFlag `json:"feature_flags"`
	}
	if err := json.Unmarshal([]byte(jsonData), &featureManagement); err != nil {
		t.Fatalf("Failed to unmarshal test data: %v", err)
	}

	provider := &mockFeatureFlagProvider{featureFlags: featureManagement.FeatureFlags}
	manager, err := NewFeatureManager(provider, nil)
	if err != nil {
		t.Fatalf("Failed to create feature manager: %v", err)
	}

	// Note: These tests will use current time, so they're more like smoke tests
	// In a real scenario, you'd want to inject time for more precise testing
	t.Run("Daily recurrence flag exists", func(t *testing.T) {
		_, err := manager.IsEnabled("DailyRecurrence")
		if err != nil {
			t.Errorf("Failed to evaluate daily recurrence flag: %v", err)
		}
	})

	t.Run("Daily recurrence with limit flag exists", func(t *testing.T) {
		_, err := manager.IsEnabled("DailyRecurrenceWithLimit")
		if err != nil {
			t.Errorf("Failed to evaluate daily recurrence with limit flag: %v", err)
		}
	})
}

func TestTimeWindowFilterWithRecurrence_WeeklyPattern(t *testing.T) {
	// Define test feature flag with weekly recurrence
	jsonData := `{
        "feature_flags": [
            {
                "id": "WeeklyRecurrence",
                "description": "A feature flag active on Monday and Friday",
                "enabled": true,
                "conditions": {
                    "client_filters": [
                        {
                            "name": "Microsoft.TimeWindow",
                            "parameters": {
                                "Start": "2023-09-01T09:00:00Z",
                                "End": "2023-09-01T17:00:00Z",
                                "Recurrence": {
                                    "Pattern": {
                                        "Type": "Weekly",
                                        "Interval": 1,
                                        "DaysOfWeek": ["Monday", "Friday"],
                                        "FirstDayOfWeek": "Sunday"
                                    },
                                    "Range": {
                                        "Type": "NoEnd"
                                    }
                                }
                            }
                        }
                    ]
                }
            },
            {
                "id": "BiWeeklyRecurrence",
                "description": "A feature flag active every other week on Tuesday",
                "enabled": true,
                "conditions": {
                    "client_filters": [
                        {
                            "name": "Microsoft.TimeWindow",
                            "parameters": {
                                "Start": "2023-09-05T09:00:00Z",
                                "End": "2023-09-05T17:00:00Z",
                                "Recurrence": {
                                    "Pattern": {
                                        "Type": "Weekly",
                                        "Interval": 2,
                                        "DaysOfWeek": ["Tuesday"],
                                        "FirstDayOfWeek": "Sunday"
                                    },
                                    "Range": {
                                        "Type": "EndDate",
                                        "EndDate": "2023-12-31T23:59:59Z"
                                    }
                                }
                            }
                        }
                    ]
                }
            }
        ]
    }`

	var featureManagement struct {
		FeatureFlags []FeatureFlag `json:"feature_flags"`
	}
	if err := json.Unmarshal([]byte(jsonData), &featureManagement); err != nil {
		t.Fatalf("Failed to unmarshal test data: %v", err)
	}

	provider := &mockFeatureFlagProvider{featureFlags: featureManagement.FeatureFlags}
	manager, err := NewFeatureManager(provider, nil)
	if err != nil {
		t.Fatalf("Failed to create feature manager: %v", err)
	}

	t.Run("Weekly recurrence flag exists", func(t *testing.T) {
		_, err := manager.IsEnabled("WeeklyRecurrence")
		if err != nil {
			t.Errorf("Failed to evaluate weekly recurrence flag: %v", err)
		}
	})

	t.Run("Bi-weekly recurrence flag exists", func(t *testing.T) {
		_, err := manager.IsEnabled("BiWeeklyRecurrence")
		if err != nil {
			t.Errorf("Failed to evaluate bi-weekly recurrence flag: %v", err)
		}
	})
}

func TestTimeWindowFilterWithRecurrence_InvalidParameters(t *testing.T) {
	// Test various invalid recurrence configurations
	testCases := []struct {
		name       string
		jsonData   string
		shouldFail bool
	}{
		{
			name: "Missing Start with Recurrence",
			jsonData: `{
                "feature_flags": [{
                    "id": "MissingStart",
                    "enabled": true,
                    "conditions": {
                        "client_filters": [{
                            "name": "Microsoft.TimeWindow",
                            "parameters": {
                                "End": "2023-09-01T17:00:00Z",
                                "Recurrence": {
                                    "Pattern": {"Type": "Daily"},
                                    "Range": {"Type": "NoEnd"}
                                }
                            }
                        }]
                    }
                }]
            }`,
			shouldFail: false, // Should return false, not error
		},
		{
			name: "Invalid Pattern Type",
			jsonData: `{
                "feature_flags": [{
                    "id": "InvalidPattern",
                    "enabled": true,
                    "conditions": {
                        "client_filters": [{
                            "name": "Microsoft.TimeWindow",
                            "parameters": {
                                "Start": "2023-09-01T09:00:00Z",
                                "End": "2023-09-01T17:00:00Z",
                                "Recurrence": {
                                    "Pattern": {"Type": "Monthly"},
                                    "Range": {"Type": "NoEnd"}
                                }
                            }
                        }]
                    }
                }]
            }`,
			shouldFail: false, // Should return false, not error
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var featureManagement struct {
				FeatureFlags []FeatureFlag `json:"feature_flags"`
			}
			if err := json.Unmarshal([]byte(tc.jsonData), &featureManagement); err != nil {
				t.Fatalf("Failed to unmarshal test data: %v", err)
			}

			provider := &mockFeatureFlagProvider{featureFlags: featureManagement.FeatureFlags}
			manager, err := NewFeatureManager(provider, nil)
			if err != nil {
				t.Fatalf("Failed to create feature manager: %v", err)
			}

			result, err := manager.IsEnabled(featureManagement.FeatureFlags[0].ID)
			if tc.shouldFail && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.shouldFail && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			// Invalid recurrence should return false
			if result {
				t.Error("Expected false for invalid recurrence configuration")
			}
		})
	}
}

// TestRecurringTimeWindow_DailyPattern tests daily recurring time window
// mirroring the TypeScript test case for "DailyTimeWindow"
// A feature flag active from 18:00:00 to 20:00:00 every other day since 2024-12-10, until 2025-1-1
func TestRecurringTimeWindow_DailyPattern(t *testing.T) {
	// Create recurrence spec matching the TypeScript test
	// Start: "Tue, 10 Dec 2024 18:00:00 GMT"
	// End: "Tue, 10 Dec 2024 20:00:00 GMT"
	// Pattern: Daily with Interval 2
	// Range: EndDate "Wed, 1 Jan 2025 20:00:00 GMT"

	startTime := time.Date(2024, 12, 10, 18, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 12, 10, 20, 0, 0, 0, time.UTC)
	rangeEndDate := time.Date(2025, 1, 1, 20, 0, 0, 0, time.UTC)

	spec := &RecurrenceSpec{
		StartTime: startTime,
		Duration:  endTime.Sub(startTime).Milliseconds(),
		Pattern: RecurrencePattern{
			Type:     Daily,
			Interval: 2, // Every other day
		},
		Range: RecurrenceRange{
			Type:    EndDate,
			EndDate: &rangeEndDate,
		},
		TimezoneOffset: 0,
	}

	testCases := []struct {
		name     string
		testTime time.Time
		expected bool
		reason   string
	}{
		{
			name:     "Before start time on first day",
			testTime: time.Date(2024, 12, 10, 17, 59, 59, 0, time.UTC),
			expected: false,
			reason:   "2024-12-10T17:59:59 is before the 18:00:00 start time",
		},
		{
			name:     "Exactly at start time on first day",
			testTime: time.Date(2024, 12, 10, 18, 0, 0, 0, time.UTC),
			expected: true,
			reason:   "2024-12-10T18:00:00 is the exact start time",
		},
		{
			name:     "Within window on first day",
			testTime: time.Date(2024, 12, 10, 19, 59, 59, 0, time.UTC),
			expected: true,
			reason:   "2024-12-10T19:59:59 is within 18:00-20:00 window",
		},
		{
			name:     "After end time on first day",
			testTime: time.Date(2024, 12, 10, 20, 0, 1, 0, time.UTC),
			expected: false,
			reason:   "2024-12-10T20:00:01 is after the 20:00:00 end time",
		},
		{
			name:     "Second day (not in interval)",
			testTime: time.Date(2024, 12, 11, 18, 0, 1, 0, time.UTC),
			expected: false,
			reason:   "2024-12-11 is day 2, interval is 2, so it skips this day",
		},
		{
			name:     "Third day (in interval)",
			testTime: time.Date(2024, 12, 12, 18, 0, 1, 0, time.UTC),
			expected: true,
			reason:   "2024-12-12 is day 3, matches the interval of 2",
		},
		{
			name:     "Dec 24 (within pattern and range)",
			testTime: time.Date(2024, 12, 24, 18, 0, 1, 0, time.UTC),
			expected: true,
			reason:   "2024-12-24 is 14 days from start (7 intervals), within range",
		},
		{
			name:     "Dec 25 (not in interval)",
			testTime: time.Date(2024, 12, 25, 18, 0, 1, 0, time.UTC),
			expected: false,
			reason:   "2024-12-25 is 15 days from start, not matching interval of 2",
		},
		{
			name:     "Jan 1 within time window (last day of range)",
			testTime: time.Date(2025, 1, 1, 18, 0, 1, 0, time.UTC),
			expected: true,
			reason:   "2025-01-01T18:00:01 is 22 days from start (11 intervals), within range end date",
		},
		{
			name:     "Jan 1 after time window (at range end)",
			testTime: time.Date(2025, 1, 1, 20, 0, 1, 0, time.UTC),
			expected: false,
			reason:   "2025-01-01T20:00:01 is after the daily window end time",
		},
		{
			name:     "Jan 3 (after range end date)",
			testTime: time.Date(2025, 1, 3, 18, 0, 1, 0, time.UTC),
			expected: false,
			reason:   "2025-01-03 is beyond the range end date",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := matchRecurrence(tc.testTime, spec)
			if result != tc.expected {
				t.Errorf("Expected %v but got %v\nReason: %s\nTest time: %v",
					tc.expected, result, tc.reason, tc.testTime)
			}
		})
	}
}

// TestRecurringTimeWindow_WeeklyPattern tests weekly recurring time window
// mirroring the TypeScript test case for "WeeklyTimeWindow"
// A feature flag active from 18:00:00 to 20:00:00 every weekday since 2024-12-10, for 10 occurrences
func TestRecurringTimeWindow_WeeklyPattern(t *testing.T) {
	// Create recurrence spec matching the TypeScript test
	// Start: "Tue, 10 Dec 2024 18:00:00 GMT"
	// End: "Tue, 10 Dec 2024 20:00:00 GMT"
	// Pattern: Weekly, Interval 1, DaysOfWeek: Monday-Friday
	// Range: Numbered, 10 occurrences

	startTime := time.Date(2024, 12, 10, 18, 0, 0, 0, time.UTC) // Tuesday
	endTime := time.Date(2024, 12, 10, 20, 0, 0, 0, time.UTC)
	occurrences := 10

	spec := &RecurrenceSpec{
		StartTime: startTime,
		Duration:  endTime.Sub(startTime).Milliseconds(),
		Pattern: RecurrencePattern{
			Type:           Weekly,
			Interval:       1,
			DaysOfWeek:     []DayOfWeek{Monday, Tuesday, Wednesday, Thursday, Friday},
			FirstDayOfWeek: Sunday, // Default
		},
		Range: RecurrenceRange{
			Type:                Numbered,
			NumberOfOccurrences: &occurrences,
		},
		TimezoneOffset: 0,
	}

	testCases := []struct {
		name     string
		testTime time.Time
		expected bool
		reason   string
	}{
		{
			name:     "Before start time on Tuesday Dec 10",
			testTime: time.Date(2024, 12, 10, 17, 59, 59, 0, time.UTC),
			expected: false,
			reason:   "Before the 18:00:00 start time on first occurrence",
		},
		{
			name:     "Tuesday Dec 10 within window (occurrence 1)",
			testTime: time.Date(2024, 12, 10, 18, 0, 1, 0, time.UTC),
			expected: true,
			reason:   "First occurrence - Tuesday within 18:00-20:00",
		},
		{
			name:     "Wednesday Dec 11 within window (occurrence 2)",
			testTime: time.Date(2024, 12, 11, 18, 0, 1, 0, time.UTC),
			expected: true,
			reason:   "Second occurrence - Wednesday within 18:00-20:00",
		},
		{
			name:     "Thursday Dec 12 within window (occurrence 3)",
			testTime: time.Date(2024, 12, 12, 18, 0, 1, 0, time.UTC),
			expected: true,
			reason:   "Third occurrence - Thursday within 18:00-20:00",
		},
		{
			name:     "Friday Dec 13 within window (occurrence 4)",
			testTime: time.Date(2024, 12, 13, 18, 0, 1, 0, time.UTC),
			expected: true,
			reason:   "Fourth occurrence - Friday within 18:00-20:00",
		},
		{
			name:     "Saturday Dec 14 (not a weekday)",
			testTime: time.Date(2024, 12, 14, 18, 0, 1, 0, time.UTC),
			expected: false,
			reason:   "Saturday is not in DaysOfWeek (weekdays only)",
		},
		{
			name:     "Sunday Dec 15 (not a weekday)",
			testTime: time.Date(2024, 12, 15, 18, 0, 1, 0, time.UTC),
			expected: false,
			reason:   "Sunday is not in DaysOfWeek (weekdays only)",
		},
		{
			name:     "Monday Dec 16 within window (occurrence 5)",
			testTime: time.Date(2024, 12, 16, 18, 0, 1, 0, time.UTC),
			expected: true,
			reason:   "Fifth occurrence - Monday of second week within 18:00-20:00",
		},
		{
			name:     "Monday Dec 16 after window",
			testTime: time.Date(2024, 12, 16, 20, 0, 1, 0, time.UTC),
			expected: false,
			reason:   "After the 20:00:00 end time",
		},
		{
			name:     "Monday Dec 23 within window (occurrence 9)",
			testTime: time.Date(2024, 12, 23, 18, 0, 1, 0, time.UTC),
			expected: true,
			reason:   "Ninth occurrence - Monday of third week",
		},
		{
			name:     "Tuesday Dec 24 (occurrence 10 - last)",
			testTime: time.Date(2024, 12, 24, 18, 0, 1, 0, time.UTC),
			expected: false,
			reason:   "Tenth occurrence would be Tuesday Dec 17, so Dec 24 is beyond 10 occurrences",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := matchRecurrence(tc.testTime, spec)
			if result != tc.expected {
				t.Errorf("Expected %v but got %v\nReason: %s\nTest time: %v (day: %v)",
					tc.expected, result, tc.reason, tc.testTime, tc.testTime.Weekday())
			}
		})
	}
}

// TestRecurringTimeWindow_EdgeCases tests additional edge cases
func TestRecurringTimeWindow_EdgeCases(t *testing.T) {
	t.Run("Daily recurrence exact boundary times", func(t *testing.T) {
		startTime := time.Date(2024, 12, 10, 18, 0, 0, 0, time.UTC)
		endTime := time.Date(2024, 12, 10, 20, 0, 0, 0, time.UTC)

		spec := &RecurrenceSpec{
			StartTime:      startTime,
			Duration:       endTime.Sub(startTime).Milliseconds(),
			Pattern:        RecurrencePattern{Type: Daily, Interval: 1},
			Range:          RecurrenceRange{Type: NoEnd},
			TimezoneOffset: 0,
		}

		// Test exact start time
		if !matchRecurrence(startTime, spec) {
			t.Error("Should match at exact start time")
		}

		// Test 1 millisecond before start
		beforeStart := startTime.Add(-1 * time.Millisecond)
		if matchRecurrence(beforeStart, spec) {
			t.Error("Should not match 1ms before start time")
		}

		// Test exact end time (exclusive)
		if matchRecurrence(endTime, spec) {
			t.Error("Should not match at exact end time (end is exclusive)")
		}

		// Test 1 millisecond before end
		beforeEnd := endTime.Add(-1 * time.Millisecond)
		if !matchRecurrence(beforeEnd, spec) {
			t.Error("Should match 1ms before end time")
		}
	})

	t.Run("Weekly recurrence with FirstDayOfWeek", func(t *testing.T) {
		// Start on Sunday with Monday as first day of week
		startTime := time.Date(2024, 12, 15, 18, 0, 0, 0, time.UTC) // Sunday
		endTime := time.Date(2024, 12, 15, 20, 0, 0, 0, time.UTC)

		spec := &RecurrenceSpec{
			StartTime: startTime,
			Duration:  endTime.Sub(startTime).Milliseconds(),
			Pattern: RecurrencePattern{
				Type:           Weekly,
				Interval:       1,
				DaysOfWeek:     []DayOfWeek{Sunday, Monday},
				FirstDayOfWeek: Monday, // Week starts on Monday
			},
			Range:          RecurrenceRange{Type: NoEnd},
			TimezoneOffset: 0,
		}

		// Test Sunday (start day)
		if !matchRecurrence(startTime, spec) {
			t.Error("Should match on Sunday (start day)")
		}

		// Test next Monday
		nextMonday := time.Date(2024, 12, 16, 18, 0, 1, 0, time.UTC)
		if !matchRecurrence(nextMonday, spec) {
			t.Error("Should match on next Monday")
		}
	})
}
