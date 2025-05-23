// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package featuremanagement

import "fmt"

// validateFeatureFlag validates an individual feature flag
func validateFeatureFlag(flag FeatureFlag) error {
	if flag.ID == "" {
		return fmt.Errorf("feature flag ID is required")
	}

	// Validate conditions if present
	if flag.Conditions != nil {
		if err := validateConditions(flag.ID, flag.Conditions); err != nil {
			return err
		}
	}

	// Validate variants if present
	if len(flag.Variants) > 0 {
		if err := validateVariantsDefinition(flag.ID, flag.Variants); err != nil {
			return err
		}
	}

	// Validate allocation if present
	if flag.Allocation != nil {
		if err := validateAllocation(flag.ID, flag.Allocation); err != nil {
			return err
		}
	}

	return nil
}

func validateConditions(id string, conditions *Conditions) error {
	// Validate requirement_type field
	if conditions.RequirementType != "" &&
		conditions.RequirementType != RequirementTypeAny &&
		conditions.RequirementType != RequirementTypeAll {
		return fmt.Errorf("invalid feature flag %s: requirement_type must be 'Any' or 'All'", id)
	}

	// Validate client filters
	for i, filter := range conditions.ClientFilters {
		if filter.Name == "" {
			return fmt.Errorf("invalid feature flag %s: client filter at index %d missing name", id, i)
		}
	}

	return nil
}

func validateVariantsDefinition(id string, variants []VariantDefinition) error {
	for i, variant := range variants {
		if variant.Name == "" {
			return fmt.Errorf("invalid feature flag %s: variant at index %d missing name", id, i)
		}

		if variant.StatusOverride != "" &&
			variant.StatusOverride != StatusOverrideNone &&
			variant.StatusOverride != StatusOverrideEnabled &&
			variant.StatusOverride != StatusOverrideDisabled {
			return fmt.Errorf("invalid feature flag %s at index %d: variant status_override must be 'None', 'Enabled', or 'Disabled'", id, i)
		}
	}

	return nil
}

func validateAllocation(id string, allocation *VariantAllocation) error {
	// Validate percentile allocations
	for i, p := range allocation.Percentile {
		if p.Variant == "" {
			return fmt.Errorf("invalid feature flag %s: percentile allocation at index %d missing variant", id, i)
		}

		if p.From < 0 || p.From > 100 {
			return fmt.Errorf("invalid feature flag %s: percentile 'from' must be between 0 and 100", id)
		}

		if p.To < 0 || p.To > 100 {
			return fmt.Errorf("invalid feature flag %s: percentile 'to' must be between 0 and 100", id)
		}
	}

	// Similar validations for user and group allocations
	for i, u := range allocation.User {
		if u.Variant == "" {
			return fmt.Errorf("invalid feature flag %s: user allocation at index %d missing variant", id, i)
		}

		if len(u.Users) == 0 {
			return fmt.Errorf("invalid feature flag %s: user allocation at index %d has empty users list", id, i)
		}
	}

	for i, g := range allocation.Group {
		if g.Variant == "" {
			return fmt.Errorf("invalid feature flag %s: group allocation at index %d missing variant", id, i)
		}

		if len(g.Groups) == 0 {
			return fmt.Errorf("invalid feature flag %s: group allocation at index %d has empty groups list", id, i)
		}
	}

	return nil
}
