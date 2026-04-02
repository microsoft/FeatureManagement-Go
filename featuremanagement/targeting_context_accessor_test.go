// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package featuremanagement

import (
	"fmt"
	"testing"

	"github.com/go-viper/mapstructure/v2"
)

// mockTargetingContextAccessor implements TargetingContextAccessor for testing
type mockTargetingContextAccessor struct {
	targetingContext TargetingContext
	err             error
}

func (m *mockTargetingContextAccessor) GetTargetingContext() (TargetingContext, error) {
	return m.targetingContext, m.err
}

func TestTargetingContextAccessor_IsEnabled(t *testing.T) {
	featureFlagData := map[string]any{
		"ID":      "TargetedFeature",
		"Enabled": true,
		"Conditions": map[string]any{
			"ClientFilters": []any{
				map[string]any{
					"Name": "Microsoft.Targeting",
					"Parameters": map[string]any{
						"Audience": map[string]any{
							"Users":                    []any{"Alice"},
							"Groups":                   []any{},
							"DefaultRolloutPercentage": 0,
						},
					},
				},
			},
		},
	}

	var featureFlag FeatureFlag
	err := mapstructure.Decode(featureFlagData, &featureFlag)
	if err != nil {
		t.Fatalf("Failed to decode feature flag: %v", err)
	}

	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{featureFlag},
	}

	tests := []struct {
		name           string
		accessor       *mockTargetingContextAccessor
		expectedResult bool
		expectError    bool
	}{
		{
			name: "Accessor provides targeted user - should be enabled",
			accessor: &mockTargetingContextAccessor{
				targetingContext: TargetingContext{UserID: "Alice"},
			},
			expectedResult: true,
		},
		{
			name: "Accessor provides non-targeted user - should be disabled",
			accessor: &mockTargetingContextAccessor{
				targetingContext: TargetingContext{UserID: "Bob"},
			},
			expectedResult: false,
		},
		{
			name: "Accessor returns error - targeting filter fails gracefully",
			accessor: &mockTargetingContextAccessor{
				err: fmt.Errorf("no user context available"),
			},
			expectedResult: false,
			expectError:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manager, err := NewFeatureManager(provider, &Options{
				TargetingContextAccessor: tc.accessor,
			})
			if err != nil {
				t.Fatalf("Failed to create feature manager: %v", err)
			}

			// Call IsEnabled without appContext — accessor should provide targeting context
			result, err := manager.IsEnabled("TargetedFeature")
			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if result != tc.expectedResult {
				t.Errorf("Expected %v, got %v", tc.expectedResult, result)
			}
		})
	}
}

func TestTargetingContextAccessor_ExplicitContextOverridesAccessor(t *testing.T) {
	featureFlagData := map[string]any{
		"ID":      "TargetedFeature",
		"Enabled": true,
		"Conditions": map[string]any{
			"ClientFilters": []any{
				map[string]any{
					"Name": "Microsoft.Targeting",
					"Parameters": map[string]any{
						"Audience": map[string]any{
							"Users":                    []any{"Alice"},
							"Groups":                   []any{},
							"DefaultRolloutPercentage": 0,
						},
					},
				},
			},
		},
	}

	var featureFlag FeatureFlag
	err := mapstructure.Decode(featureFlagData, &featureFlag)
	if err != nil {
		t.Fatalf("Failed to decode feature flag: %v", err)
	}

	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{featureFlag},
	}

	// Accessor returns "Bob" (not targeted), but explicit context says "Alice" (targeted)
	accessor := &mockTargetingContextAccessor{
		targetingContext: TargetingContext{UserID: "Bob"},
	}

	manager, err := NewFeatureManager(provider, &Options{
		TargetingContextAccessor: accessor,
	})
	if err != nil {
		t.Fatalf("Failed to create feature manager: %v", err)
	}

	// Explicit context should override the accessor
	result, err := manager.IsEnabledWithAppContext("TargetedFeature", TargetingContext{UserID: "Alice"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}
	if !result {
		t.Error("Expected feature to be enabled for Alice (explicit context), but it was disabled")
	}
}

func TestTargetingContextAccessor_GetVariant(t *testing.T) {
	featureFlagData := map[string]any{
		"ID":      "VariantFeature",
		"Enabled": true,
		"Allocation": map[string]any{
			"DefaultWhenEnabled": "Small",
			"User": []any{
				map[string]any{
					"Variant": "Big",
					"Users":   []any{"Alice"},
				},
			},
		},
		"Variants": []any{
			map[string]any{"Name": "Big", "ConfigurationValue": "500px"},
			map[string]any{"Name": "Small", "ConfigurationValue": "300px"},
		},
	}

	var featureFlag FeatureFlag
	err := mapstructure.Decode(featureFlagData, &featureFlag)
	if err != nil {
		t.Fatalf("Failed to decode feature flag: %v", err)
	}

	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{featureFlag},
	}

	accessor := &mockTargetingContextAccessor{
		targetingContext: TargetingContext{UserID: "Alice"},
	}

	manager, err := NewFeatureManager(provider, &Options{
		TargetingContextAccessor: accessor,
	})
	if err != nil {
		t.Fatalf("Failed to create feature manager: %v", err)
	}

	// Call GetVariant with nil appContext — accessor should provide targeting context
	variant, err := manager.GetVariant("VariantFeature", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if variant == nil {
		t.Fatal("Expected a variant but got nil")
	}

	if variant.Name != "Big" {
		t.Errorf("Expected variant 'Big' for Alice, got '%s'", variant.Name)
	}
}

func TestTargetingContextAccessor_NilAccessor(t *testing.T) {
	featureFlagData := map[string]any{
		"ID":      "SimpleFeature",
		"Enabled": true,
	}

	var featureFlag FeatureFlag
	err := mapstructure.Decode(featureFlagData, &featureFlag)
	if err != nil {
		t.Fatalf("Failed to decode feature flag: %v", err)
	}

	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{featureFlag},
	}

	// No accessor — should work fine for features without targeting
	manager, err := NewFeatureManager(provider, nil)
	if err != nil {
		t.Fatalf("Failed to create feature manager: %v", err)
	}

	result, err := manager.IsEnabled("SimpleFeature")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}
	if !result {
		t.Error("Expected feature to be enabled")
	}
}
