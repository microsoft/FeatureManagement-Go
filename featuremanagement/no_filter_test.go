package featuremanagement

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestFeatureManager_IsEnabled(t *testing.T) {
	// Create the mock provider
	provider := &mockFeatureFlagProvider{
		featureFlags: createTestFeatureFlags(),
	}

	// Create a feature manager with the mock provider
	fm, err := NewFeatureManager(provider, nil)
	if err != nil {
		t.Fatalf("Failed to create feature manager: %v", err)
	}

	// Test cases
	tests := []struct {
		name           string
		featureName    string
		expectedResult bool
		expectError    bool
	}{
		{
			name:           "BooleanTrue should be enabled",
			featureName:    "BooleanTrue",
			expectedResult: true,
			expectError:    false,
		},
		{
			name:           "BooleanFalse should be disabled",
			featureName:    "BooleanFalse",
			expectedResult: false,
			expectError:    false,
		},
		{
			name:           "Minimal should be enabled",
			featureName:    "Minimal",
			expectedResult: true,
			expectError:    false,
		},
		{
			name:           "NoEnabled should be disabled",
			featureName:    "NoEnabled",
			expectedResult: false,
			expectError:    false,
		},
		{
			name:           "EmptyConditions should be enabled",
			featureName:    "EmptyConditions",
			expectedResult: true,
			expectError:    false,
		},
		{
			name:           "NonExistentFeature should error",
			featureName:    "NonExistentFeature",
			expectedResult: false,
			expectError:    true,
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := fm.IsEnabled(tc.featureName)

			// Check error expectations
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Did not expect error but got: %v", err)
			}

			// If we're not expecting an error, check the result
			if !tc.expectError && result != tc.expectedResult {
				t.Errorf("Expected result %v, got %v", tc.expectedResult, result)
			}
		})
	}
}

// Create mock feature flags from the provided JSON
func createTestFeatureFlags() []FeatureFlag {
	jsonData := `{
        "feature_flags": [
            {
                "id": "BooleanTrue",
                "description": "A feature flag with no Filters, that returns true.",
                "enabled": true,
                "conditions": {
                    "client_filters": []
                }
            },
            {
                "id": "BooleanFalse",
                "description": "A feature flag with no Filters, that returns false.",
                "enabled": false,
                "conditions": {
                    "client_filters": []
                }
            },
            {
                "id": "Minimal",
                "enabled": true
            },
            {
                "id": "NoEnabled"
            },
            {
                "id": "EmptyConditions",
                "description": "A feature flag with no values in conditions, that returns true.",
                "enabled": true,
                "conditions": {}
            }
        ]
    }`

	var featureManagement struct {
		FeatureFlags []FeatureFlag `json:"feature_flags"`
	}

	if err := json.Unmarshal([]byte(jsonData), &featureManagement); err != nil {
		panic("Failed to unmarshal test feature flags: " + err.Error())
	}

	return featureManagement.FeatureFlags
}

// Mock feature flag provider for testing
type mockFeatureFlagProvider struct {
	featureFlags []FeatureFlag
}

func (m *mockFeatureFlagProvider) GetFeatureFlag(name string) (FeatureFlag, error) {
	for _, flag := range m.featureFlags {
		if flag.ID == name {
			return flag, nil
		}
	}
	return FeatureFlag{}, fmt.Errorf("feature flag '%s' not found", name)
}

func (m *mockFeatureFlagProvider) GetFeatureFlags() ([]FeatureFlag, error) {
	return m.featureFlags, nil
}

func TestInvalidEnabledFeatureFlag(t *testing.T) {
	// Raw JSON with invalid enabled type
	jsonData := `{
        "id": "InvalidEnabled",
        "description": "A feature flag with an invalid 'enabled' value, that throws an exception.",
        "enabled": "invalid",
        "conditions": {
            "client_filters": []
        }
    }`

	// Try to unmarshal directly to see the error
	var flag FeatureFlag
	err := json.Unmarshal([]byte(jsonData), &flag)

	if err == nil {
		t.Error("Expected error when unmarshaling invalid enabled value, but got none")
	} else {
		t.Logf("Got expected error: %v", err)
	}
}
