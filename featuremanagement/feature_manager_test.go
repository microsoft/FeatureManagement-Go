// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package featuremanagement

import (
	"encoding/json"
	"fmt"
	"testing"
)

// errorFilter is a custom filter that always returns an error
type errorFilter struct{}

func (f *errorFilter) Name() string { return "ErrorFilter" }
func (f *errorFilter) Evaluate(_ FeatureFilterEvaluationContext, _ any) (bool, error) {
	return false, fmt.Errorf("filter error")
}

// errorProvider is a mock provider that always returns errors
type errorProvider struct{}

func (p *errorProvider) GetFeatureFlag(name string) (FeatureFlag, error) {
	return FeatureFlag{}, fmt.Errorf("provider error")
}
func (p *errorProvider) GetFeatureFlags() ([]FeatureFlag, error) {
	return nil, fmt.Errorf("provider error")
}

func TestNewFeatureManager_NilProvider(t *testing.T) {
	_, err := NewFeatureManager(nil, nil)
	if err == nil {
		t.Error("Expected error when provider is nil, got nil")
	}
}

func TestNewFeatureManager_CustomFilters(t *testing.T) {
	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{
			{
				ID:      "CustomFilterFeature",
				Enabled: true,
				Conditions: &Conditions{
					ClientFilters: []ClientFilter{
						{Name: "AlwaysTrue"},
					},
				},
			},
		},
	}

	fm, err := NewFeatureManager(provider, &Options{
		Filters: []FeatureFilter{&alwaysTrueFilter{}},
	})
	if err != nil {
		t.Fatalf("Failed to create feature manager: %v", err)
	}

	enabled, err := fm.IsEnabled("CustomFilterFeature")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !enabled {
		t.Error("Expected feature to be enabled with custom AlwaysTrue filter")
	}
}

func TestNewFeatureManager_NilFilterInOptions(t *testing.T) {
	provider := &mockFeatureFlagProvider{}
	_, err := NewFeatureManager(provider, &Options{
		Filters: []FeatureFilter{nil},
	})
	if err != nil {
		t.Fatalf("Should not error with nil filter in options: %v", err)
	}
}

func TestGetFeatureNames(t *testing.T) {
	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{
			{ID: "Feature1", Enabled: true},
			{ID: "Feature2", Enabled: false},
			{ID: "Feature3", Enabled: true},
		},
	}

	fm, err := NewFeatureManager(provider, nil)
	if err != nil {
		t.Fatalf("Failed to create feature manager: %v", err)
	}

	names := fm.GetFeatureNames()
	if len(names) != 3 {
		t.Fatalf("Expected 3 feature names, got %d", len(names))
	}

	expected := []string{"Feature1", "Feature2", "Feature3"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("Expected name %q at index %d, got %q", expected[i], i, name)
		}
	}
}

func TestGetFeatureNames_Empty(t *testing.T) {
	provider := &mockFeatureFlagProvider{featureFlags: []FeatureFlag{}}
	fm, _ := NewFeatureManager(provider, nil)

	names := fm.GetFeatureNames()
	if len(names) != 0 {
		t.Errorf("Expected 0 feature names, got %d", len(names))
	}
}

func TestGetFeatureNames_ProviderError(t *testing.T) {
	fm, _ := NewFeatureManager(&errorProvider{}, nil)

	names := fm.GetFeatureNames()
	if names != nil {
		t.Errorf("Expected nil when provider errors, got %v", names)
	}
}

func TestUnknownFilter(t *testing.T) {
	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{
			{
				ID:      "UnknownFilterFeature",
				Enabled: true,
				Conditions: &Conditions{
					ClientFilters: []ClientFilter{
						{Name: "NonExistentFilter"},
					},
				},
			},
		},
	}

	fm, _ := NewFeatureManager(provider, nil)

	enabled, err := fm.IsEnabled("UnknownFilterFeature")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if enabled {
		t.Error("Expected feature with unknown filter to be disabled")
	}
}

func TestRequirementTypeAll(t *testing.T) {
	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{
			{
				ID:      "AllFiltersTrue",
				Enabled: true,
				Conditions: &Conditions{
					RequirementType: RequirementTypeAll,
					ClientFilters: []ClientFilter{
						{Name: "AlwaysTrue"},
						{Name: "AlwaysTrue"},
					},
				},
			},
			{
				ID:      "AllFiltersMixed",
				Enabled: true,
				Conditions: &Conditions{
					RequirementType: RequirementTypeAll,
					ClientFilters: []ClientFilter{
						{Name: "AlwaysTrue"},
						{Name: "AlwaysFalse"},
					},
				},
			},
		},
	}

	fm, _ := NewFeatureManager(provider, &Options{
		Filters: []FeatureFilter{&alwaysTrueFilter{}, &alwaysFalseFilter{}},
	})

	t.Run("All filters true", func(t *testing.T) {
		enabled, err := fm.IsEnabled("AllFiltersTrue")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !enabled {
			t.Error("Expected feature to be enabled when all filters return true")
		}
	})

	t.Run("Mixed filters with RequirementTypeAll", func(t *testing.T) {
		enabled, err := fm.IsEnabled("AllFiltersMixed")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if enabled {
			t.Error("Expected feature to be disabled when one filter returns false with RequirementTypeAll")
		}
	})
}

func TestRequirementTypeAny(t *testing.T) {
	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{
			{
				ID:      "AnyFilterMixed",
				Enabled: true,
				Conditions: &Conditions{
					RequirementType: RequirementTypeAny,
					ClientFilters: []ClientFilter{
						{Name: "AlwaysFalse"},
						{Name: "AlwaysTrue"},
					},
				},
			},
			{
				ID:      "AnyFilterAllFalse",
				Enabled: true,
				Conditions: &Conditions{
					RequirementType: RequirementTypeAny,
					ClientFilters: []ClientFilter{
						{Name: "AlwaysFalse"},
						{Name: "AlwaysFalse"},
					},
				},
			},
		},
	}

	fm, _ := NewFeatureManager(provider, &Options{
		Filters: []FeatureFilter{&alwaysTrueFilter{}, &alwaysFalseFilter{}},
	})

	t.Run("Any filter true", func(t *testing.T) {
		enabled, _ := fm.IsEnabled("AnyFilterMixed")
		if !enabled {
			t.Error("Expected feature to be enabled when one filter returns true with RequirementTypeAny")
		}
	})

	t.Run("All filters false", func(t *testing.T) {
		enabled, _ := fm.IsEnabled("AnyFilterAllFalse")
		if enabled {
			t.Error("Expected feature to be disabled when all filters return false")
		}
	})
}

func TestFilterError(t *testing.T) {
	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{
			{
				ID:      "ErrorFilterFeature",
				Enabled: true,
				Conditions: &Conditions{
					ClientFilters: []ClientFilter{
						{Name: "ErrorFilter"},
					},
				},
			},
		},
	}

	fm, _ := NewFeatureManager(provider, &Options{
		Filters: []FeatureFilter{&errorFilter{}},
	})

	_, err := fm.IsEnabled("ErrorFilterFeature")
	if err == nil {
		t.Error("Expected error from filter evaluation")
	}
}

func TestStatusOverrideEnabled(t *testing.T) {
	jsonData := `{
		"feature_flags": [
			{
				"id": "StatusOverrideEnabledFeature",
				"enabled": true,
				"variants": [
					{
						"name": "MyVariant",
						"configuration_value": "override-value",
						"status_override": "Enabled"
					}
				],
				"allocation": {
					"user": [
						{
							"variant": "MyVariant",
							"users": ["TestUser"]
						}
					]
				}
			}
		]
	}`

	var fm struct {
		FeatureFlags []FeatureFlag `json:"feature_flags"`
	}
	json.Unmarshal([]byte(jsonData), &fm)

	provider := &mockFeatureFlagProvider{featureFlags: fm.FeatureFlags}
	manager, _ := NewFeatureManager(provider, nil)

	ctx := TargetingContext{UserID: "TestUser"}

	enabled, err := manager.IsEnabledWithAppContext("StatusOverrideEnabledFeature", ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !enabled {
		t.Error("Expected feature to be enabled due to StatusOverride=Enabled")
	}
}

func TestGetVariant_NonExistentFeature(t *testing.T) {
	provider := &mockFeatureFlagProvider{featureFlags: []FeatureFlag{}}
	manager, _ := NewFeatureManager(provider, nil)

	variant, err := manager.GetVariant("NonExistent", nil)
	if err == nil {
		t.Error("Expected error for non-existent feature")
	}
	if variant != nil {
		t.Error("Expected nil variant for non-existent feature")
	}
}

func TestGetVariant_NoVariants(t *testing.T) {
	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{
			{ID: "NoVariants", Enabled: true},
		},
	}
	manager, _ := NewFeatureManager(provider, nil)

	variant, err := manager.GetVariant("NoVariants", TargetingContext{UserID: "user1"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if variant != nil {
		t.Error("Expected nil variant when feature has no variants")
	}
}

func TestGetVariant_NoAllocation(t *testing.T) {
	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{
			{
				ID:      "NoAllocation",
				Enabled: true,
				Variants: []VariantDefinition{
					{Name: "Small", ConfigurationValue: "300px"},
				},
			},
		},
	}
	manager, _ := NewFeatureManager(provider, nil)

	variant, err := manager.GetVariant("NoAllocation", TargetingContext{UserID: "user1"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if variant != nil {
		t.Error("Expected nil variant when no allocation matches")
	}
}

func TestGetVariant_UserOverridesDefault(t *testing.T) {
	jsonData := `{
		"feature_flags": [
			{
				"id": "UserOverride",
				"enabled": true,
				"variants": [
					{"name": "Medium", "configuration_value": "450px"},
					{"name": "Small", "configuration_value": "300px"}
				],
				"allocation": {
					"default_when_enabled": "Medium",
					"user": [
						{"variant": "Small", "users": ["Jeff"]}
					]
				}
			}
		]
	}`

	var fm struct {
		FeatureFlags []FeatureFlag `json:"feature_flags"`
	}
	json.Unmarshal([]byte(jsonData), &fm)

	provider := &mockFeatureFlagProvider{featureFlags: fm.FeatureFlags}
	manager, _ := NewFeatureManager(provider, nil)

	variant, err := manager.GetVariant("UserOverride", TargetingContext{UserID: "Jeff"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if variant == nil {
		t.Fatal("Expected variant, got nil")
	}
	if variant.Name != "Small" {
		t.Errorf("Expected user-allocated variant 'Small', got '%s'", variant.Name)
	}
}

func TestGetVariant_PointerTargetingContext(t *testing.T) {
	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{
			{
				ID:      "PtrCtx",
				Enabled: true,
				Variants: []VariantDefinition{
					{Name: "V1", ConfigurationValue: "val"},
				},
				Allocation: &VariantAllocation{
					User: []UserAllocation{
						{Variant: "V1", Users: []string{"user1"}},
					},
				},
			},
		},
	}
	manager, _ := NewFeatureManager(provider, nil)

	ctx := &TargetingContext{UserID: "user1"}
	variant, err := manager.GetVariant("PtrCtx", ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if variant == nil || variant.Name != "V1" {
		t.Error("Expected variant V1 when using pointer TargetingContext")
	}
}

func TestIsEnabledWithAppContext_NonTargetingContext(t *testing.T) {
	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{
			{ID: "SimpleEnabled", Enabled: true},
		},
	}
	manager, _ := NewFeatureManager(provider, nil)

	enabled, err := manager.IsEnabledWithAppContext("SimpleEnabled", "some-string-context")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !enabled {
		t.Error("Expected feature to be enabled")
	}
}

// Validation tests

func TestValidation_EmptyID(t *testing.T) {
	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{
			{ID: "", Enabled: true},
		},
	}
	manager, _ := NewFeatureManager(provider, nil)

	_, err := manager.IsEnabled("")
	if err == nil {
		t.Error("Expected validation error for empty feature ID")
	}
}

func TestValidation_InvalidRequirementType(t *testing.T) {
	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{
			{
				ID:      "BadReqType",
				Enabled: true,
				Conditions: &Conditions{
					RequirementType: "Invalid",
				},
			},
		},
	}
	manager, _ := NewFeatureManager(provider, nil)

	_, err := manager.IsEnabled("BadReqType")
	if err == nil {
		t.Error("Expected validation error for invalid requirement type")
	}
}

func TestValidation_VariantMissingName(t *testing.T) {
	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{
			{
				ID:       "BadVariant",
				Enabled:  true,
				Variants: []VariantDefinition{{Name: ""}},
			},
		},
	}
	manager, _ := NewFeatureManager(provider, nil)

	_, err := manager.IsEnabled("BadVariant")
	if err == nil {
		t.Error("Expected validation error for variant missing name")
	}
}

func TestValidation_InvalidStatusOverride(t *testing.T) {
	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{
			{
				ID:       "BadOverride",
				Enabled:  true,
				Variants: []VariantDefinition{{Name: "V1", StatusOverride: "Invalid"}},
			},
		},
	}
	manager, _ := NewFeatureManager(provider, nil)

	_, err := manager.IsEnabled("BadOverride")
	if err == nil {
		t.Error("Expected validation error for invalid status override")
	}
}

func TestValidation_InvalidPercentileRange(t *testing.T) {
	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{
			{
				ID:      "BadPercentile",
				Enabled: true,
				Allocation: &VariantAllocation{
					Percentile: []PercentileAllocation{
						{Variant: "V1", From: -1, To: 50},
					},
				},
			},
		},
	}
	manager, _ := NewFeatureManager(provider, nil)

	_, err := manager.IsEnabled("BadPercentile")
	if err == nil {
		t.Error("Expected validation error for invalid percentile range")
	}
}

func TestValidation_UserAllocationEmptyUsers(t *testing.T) {
	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{
			{
				ID:      "EmptyUsers",
				Enabled: true,
				Allocation: &VariantAllocation{
					User: []UserAllocation{
						{Variant: "V1", Users: []string{}},
					},
				},
			},
		},
	}
	manager, _ := NewFeatureManager(provider, nil)

	_, err := manager.IsEnabled("EmptyUsers")
	if err == nil {
		t.Error("Expected validation error for empty users list")
	}
}

func TestValidation_GroupAllocationEmptyGroups(t *testing.T) {
	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{
			{
				ID:      "EmptyGroups",
				Enabled: true,
				Allocation: &VariantAllocation{
					Group: []GroupAllocation{
						{Variant: "V1", Groups: []string{}},
					},
				},
			},
		},
	}
	manager, _ := NewFeatureManager(provider, nil)

	_, err := manager.IsEnabled("EmptyGroups")
	if err == nil {
		t.Error("Expected validation error for empty groups list")
	}
}

func TestValidation_FilterMissingName(t *testing.T) {
	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{
			{
				ID:      "EmptyFilterName",
				Enabled: true,
				Conditions: &Conditions{
					ClientFilters: []ClientFilter{
						{Name: ""},
					},
				},
			},
		},
	}
	manager, _ := NewFeatureManager(provider, nil)

	_, err := manager.IsEnabled("EmptyFilterName")
	if err == nil {
		t.Error("Expected validation error for filter missing name")
	}
}

func TestValidation_PercentileAllocationMissingVariant(t *testing.T) {
	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{
			{
				ID:      "NoVariantPercentile",
				Enabled: true,
				Allocation: &VariantAllocation{
					Percentile: []PercentileAllocation{
						{Variant: "", From: 0, To: 50},
					},
				},
			},
		},
	}
	manager, _ := NewFeatureManager(provider, nil)

	_, err := manager.IsEnabled("NoVariantPercentile")
	if err == nil {
		t.Error("Expected validation error for percentile allocation missing variant")
	}
}

func TestValidation_UserAllocationMissingVariant(t *testing.T) {
	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{
			{
				ID:      "NoVariantUser",
				Enabled: true,
				Allocation: &VariantAllocation{
					User: []UserAllocation{
						{Variant: "", Users: []string{"user1"}},
					},
				},
			},
		},
	}
	manager, _ := NewFeatureManager(provider, nil)

	_, err := manager.IsEnabled("NoVariantUser")
	if err == nil {
		t.Error("Expected validation error for user allocation missing variant")
	}
}

func TestValidation_GroupAllocationMissingVariant(t *testing.T) {
	provider := &mockFeatureFlagProvider{
		featureFlags: []FeatureFlag{
			{
				ID:      "NoVariantGroup",
				Enabled: true,
				Allocation: &VariantAllocation{
					Group: []GroupAllocation{
						{Variant: "", Groups: []string{"group1"}},
					},
				},
			},
		},
	}
	manager, _ := NewFeatureManager(provider, nil)

	_, err := manager.IsEnabled("NoVariantGroup")
	if err == nil {
		t.Error("Expected validation error for group allocation missing variant")
	}
}

func TestIsEnabled_ProviderError(t *testing.T) {
	manager, _ := NewFeatureManager(&errorProvider{}, nil)

	_, err := manager.IsEnabled("AnyFeature")
	if err == nil {
		t.Error("Expected error when provider fails")
	}
}

func TestIsEnabledWithAppContext_ProviderError(t *testing.T) {
	manager, _ := NewFeatureManager(&errorProvider{}, nil)

	_, err := manager.IsEnabledWithAppContext("AnyFeature", nil)
	if err == nil {
		t.Error("Expected error when provider fails")
	}
}

func TestGetVariant_ProviderError(t *testing.T) {
	manager, _ := NewFeatureManager(&errorProvider{}, nil)

	_, err := manager.GetVariant("AnyFeature", nil)
	if err == nil {
		t.Error("Expected error when provider fails")
	}
}

func TestFeatureManagementStruct(t *testing.T) {
	jsonData := `{
		"feature_flags": [
			{"id": "F1", "enabled": true},
			{"id": "F2", "enabled": false}
		]
	}`

	var fm FeatureManagement
	err := json.Unmarshal([]byte(jsonData), &fm)
	if err != nil {
		t.Fatalf("Failed to unmarshal FeatureManagement: %v", err)
	}
	if len(fm.FeatureFlags) != 2 {
		t.Errorf("Expected 2 feature flags, got %d", len(fm.FeatureFlags))
	}
	if fm.FeatureFlags[0].ID != "F1" {
		t.Errorf("Expected first flag ID 'F1', got '%s'", fm.FeatureFlags[0].ID)
	}
}
