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
                                "Start": "2023-06-29T07:00:00Z",
                                "End": "2023-08-30T07:00:00Z"
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
                                "Start": "3023-06-27T06:00:00Z",
                                "End": "3023-06-28T06:05:00Z"
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
                                "Start": "2023-06-29T07:00:00Z",
                                "End": "3023-06-28T06:05:00Z"
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
	manager, err := NewFeatureManager(provider, []FeatureFilter{&TimeWindowFilter{}})
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
