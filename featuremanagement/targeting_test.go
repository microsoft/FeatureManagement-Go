// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package featuremanagement

import (
	"testing"

	"github.com/go-viper/mapstructure/v2"
)

func TestTargetingFilter(t *testing.T) {
	featureFlagData := map[string]any{
		"ID":          "ComplexTargeting",
		"Description": "A feature flag using a targeting filter, that will return true for Alice, Stage1, and 50% of Stage2. Dave and Stage3 are excluded. The default rollout percentage is 25%.",
		"Enabled":     true,
		"Conditions": map[string]any{
			"ClientFilters": []any{
				map[string]any{
					"Name": "Microsoft.Targeting",
					"Parameters": map[string]any{
						"Audience": map[string]any{
							"Users": []any{"Alice"},
							"Groups": []any{
								map[string]any{
									"Name":              "Stage1",
									"RolloutPercentage": 100,
								},
								map[string]any{
									"Name":              "Stage2",
									"RolloutPercentage": 50,
								},
							},
							"DefaultRolloutPercentage": 25,
							"Exclusion": map[string]any{
								"Users":  []any{"Dave"},
								"Groups": []any{"Stage3"},
							},
						},
					},
				},
			},
		},
	}

	var featureFlag FeatureFlag
	err := mapstructure.Decode(featureFlagData, &featureFlag)
	if err != nil {
		t.Fatalf("Failed to parse feature flag JSON: %v", err)
	}

	// Create a mock provider
	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{featureFlag},
	}

	// Create the feature manager with targeting filter
	manager, err := NewFeatureManager(provider, nil)
	if err != nil {
		t.Fatalf("Failed to create feature manager: %v", err)
	}

	// Test cases
	tests := []struct {
		name           string
		userId         string
		groups         []string
		expectedResult bool
		explanation    string
	}{
		{
			name:           "Aiden not in default rollout",
			userId:         "Aiden",
			groups:         nil,
			expectedResult: false,
			explanation:    "Aiden is not in the 25% default rollout",
		},
		{
			name:           "Blossom in default rollout",
			userId:         "Blossom",
			groups:         nil,
			expectedResult: true,
			explanation:    "Blossom is in the 25% default rollout",
		},
		{
			name:           "Alice directly targeted",
			userId:         "Alice",
			groups:         nil,
			expectedResult: true,
			explanation:    "Alice is directly targeted",
		},
		{
			name:           "Aiden in Stage1 (100% rollout)",
			userId:         "Aiden",
			groups:         []string{"Stage1"},
			expectedResult: true,
			explanation:    "Aiden is in because Stage1 is 100% rollout",
		},
		{
			name:           "Empty user in Stage2 (50% rollout)",
			userId:         "",
			groups:         []string{"Stage2"},
			expectedResult: false,
			explanation:    "Empty user is not in the 50% rollout of group Stage2",
		},
		{
			name:           "Aiden in Stage2 (50% rollout)",
			userId:         "Aiden",
			groups:         []string{"Stage2"},
			expectedResult: true,
			explanation:    "Aiden is in the 50% rollout of group Stage2",
		},
		{
			name:           "Chris in Stage2 (50% rollout)",
			userId:         "Chris",
			groups:         []string{"Stage2"},
			expectedResult: false,
			explanation:    "Chris is not in the 50% rollout of group Stage2",
		},
		{
			name:           "Stage3 group excluded",
			userId:         "",
			groups:         []string{"Stage3"},
			expectedResult: false,
			explanation:    "Stage3 group is excluded",
		},
		{
			name:           "Alice in Stage3 (excluded group)",
			userId:         "Alice",
			groups:         []string{"Stage3"},
			expectedResult: false,
			explanation:    "Alice is excluded because she is part of Stage3 group",
		},
		{
			name:           "Blossom in Stage3 (excluded group)",
			userId:         "Blossom",
			groups:         []string{"Stage3"},
			expectedResult: false,
			explanation:    "Blossom is excluded because she is part of Stage3 group",
		},
		{
			name:           "Dave in Stage1 (excluded user)",
			userId:         "Dave",
			groups:         []string{"Stage1"},
			expectedResult: false,
			explanation:    "Dave is excluded because he is in the exclusion list",
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create targeting context
			targetingContext := TargetingContext{
				UserID: tc.userId,
				Groups: tc.groups,
			}

			// Evaluate the feature flag
			result, err := manager.IsEnabledWithAppContext("ComplexTargeting", targetingContext)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tc.expectedResult {
				t.Errorf("Expected %v, got %v - %s", tc.expectedResult, result, tc.explanation)
			}
		})
	}
}
