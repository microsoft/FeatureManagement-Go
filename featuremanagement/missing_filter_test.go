// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package featuremanagement

import (
	"testing"
)

// alwaysTrueFilter is a test filter that always returns true.
type alwaysTrueFilter struct{}

func (f *alwaysTrueFilter) Name() string { return "AlwaysTrue" }
func (f *alwaysTrueFilter) Evaluate(_ FeatureFilterEvaluationContext, _ any) (bool, error) {
	return true, nil
}

// alwaysFalseFilter is a test filter that always returns false.
type alwaysFalseFilter struct{}

func (f *alwaysFalseFilter) Name() string { return "AlwaysFalse" }
func (f *alwaysFalseFilter) Evaluate(_ FeatureFilterEvaluationContext, _ any) (bool, error) {
	return false, nil
}

func TestMissingFilter_RequirementTypeAny(t *testing.T) {
	tests := []struct {
		name           string
		filters        []ClientFilter
		expectedResult bool
		explanation    string
	}{
		{
			name: "Missing filter followed by matching filter should be enabled",
			filters: []ClientFilter{
				{Name: "UnregisteredFilter"},
				{Name: "AlwaysTrue"},
			},
			expectedResult: true,
			explanation:    "With RequirementType Any, a missing filter should be skipped and the matching AlwaysTrue filter should enable the feature",
		},
		{
			name: "Matching filter followed by missing filter should be enabled",
			filters: []ClientFilter{
				{Name: "AlwaysTrue"},
				{Name: "UnregisteredFilter"},
			},
			expectedResult: true,
			explanation:    "With RequirementType Any, AlwaysTrue matches first so the feature should be enabled",
		},
		{
			name: "Only missing filters should be disabled",
			filters: []ClientFilter{
				{Name: "UnregisteredFilter"},
				{Name: "AnotherUnregisteredFilter"},
			},
			expectedResult: false,
			explanation:    "With RequirementType Any, all filters are missing so no filter can match",
		},
		{
			name: "Missing filter with non-matching filter should be disabled",
			filters: []ClientFilter{
				{Name: "UnregisteredFilter"},
				{Name: "AlwaysFalse"},
			},
			expectedResult: false,
			explanation:    "With RequirementType Any, missing filter is skipped and AlwaysFalse does not match",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			provider := &mockFeatureFlagProvider{
				featureFlags: []FeatureFlag{
					{
						ID:      "TestFeature",
						Enabled: true,
						Conditions: &Conditions{
							RequirementType: RequirementTypeAny,
							ClientFilters:   tc.filters,
						},
					},
				},
			}

			fm, err := NewFeatureManager(provider, &Options{
				Filters: []FeatureFilter{&alwaysTrueFilter{}, &alwaysFalseFilter{}},
			})
			if err != nil {
				t.Fatalf("Failed to create feature manager: %v", err)
			}

			result, err := fm.IsEnabled("TestFeature")
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result != tc.expectedResult {
				t.Errorf("Expected %v, got %v - %s", tc.expectedResult, result, tc.explanation)
			}
		})
	}
}

func TestMissingFilter_RequirementTypeAll(t *testing.T) {
	tests := []struct {
		name           string
		filters        []ClientFilter
		expectedResult bool
		explanation    string
	}{
		{
			name: "Missing filter with matching filter should be disabled",
			filters: []ClientFilter{
				{Name: "UnregisteredFilter"},
				{Name: "AlwaysTrue"},
			},
			expectedResult: false,
			explanation:    "With RequirementType All, a missing filter means not all filters can pass so the feature should be disabled",
		},
		{
			name: "Matching filter followed by missing filter should be disabled",
			filters: []ClientFilter{
				{Name: "AlwaysTrue"},
				{Name: "UnregisteredFilter"},
			},
			expectedResult: false,
			explanation:    "With RequirementType All, a missing filter means not all filters can pass so the feature should be disabled",
		},
		{
			name: "Only missing filters should be disabled",
			filters: []ClientFilter{
				{Name: "UnregisteredFilter"},
			},
			expectedResult: false,
			explanation:    "With RequirementType All, a missing filter means the feature should be disabled",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			provider := &mockFeatureFlagProvider{
				featureFlags: []FeatureFlag{
					{
						ID:      "TestFeature",
						Enabled: true,
						Conditions: &Conditions{
							RequirementType: RequirementTypeAll,
							ClientFilters:   tc.filters,
						},
					},
				},
			}

			fm, err := NewFeatureManager(provider, &Options{
				Filters: []FeatureFilter{&alwaysTrueFilter{}, &alwaysFalseFilter{}},
			})
			if err != nil {
				t.Fatalf("Failed to create feature manager: %v", err)
			}

			result, err := fm.IsEnabled("TestFeature")
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result != tc.expectedResult {
				t.Errorf("Expected %v, got %v - %s", tc.expectedResult, result, tc.explanation)
			}
		})
	}
}
