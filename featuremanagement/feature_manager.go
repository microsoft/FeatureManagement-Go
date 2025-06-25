// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package featuremanagement

import (
	"fmt"
	"log"
)

// EvaluationResult contains information about a feature flag evaluation
type EvaluationResult struct {
	// Feature contains the evaluated feature flag
	Feature *FeatureFlag
	// Enabled indicates the final state of the feature after evaluation
	Enabled bool
	// TargetingID is the identifier used for consistent targeting
	TargetingID string
	// Variant is the selected variant (if any)
	Variant *Variant
	// VariantAssignmentReason explains why the variant was assigned
	VariantAssignmentReason VariantAssignmentReason
}

// FeatureManager is responsible for evaluating feature flags and their variants.
// It is the main entry point for interacting with the feature management library.
type FeatureManager struct {
	featureProvider    FeatureFlagProvider
	featureFilters     map[string]FeatureFilter
	onFeatureEvaluated []func(evalRes EvaluationResult)
}

// Options configures the behavior of the FeatureManager.
type Options struct {
    // Filters is a list of custom feature filters that will be used during feature flag evaluation.
    // Each filter must implement the FeatureFilter interface.
    Filters []FeatureFilter
}

// NewFeatureManager creates and initializes a new instance of the FeatureManager.
// This is the entry point for using feature management functionality.
//
// Parameters:
//   - provider: A FeatureFlagProvider that supplies feature flag definitions
//     from a source such as Azure App Configuration or a local JSON file
//   - *options: Configuration options for the FeatureManager, including custom filters
//     for conditional feature evaluation
//
// Returns:
//   - *FeatureManager: A configured feature manager instance ready for use
//   - error: An error if initialization fails
func NewFeatureManager(provider FeatureFlagProvider, options *Options) (*FeatureManager, error) {
	if provider == nil {
		return nil, fmt.Errorf("feature provider cannot be nil")
	}

	if options == nil {
		options = &Options{}
	}

	filters := []FeatureFilter{
		&TargetingFilter{},
		&TimeWindowFilter{},
	}

	filters = append(filters, options.Filters...)
	featureFilters := make(map[string]FeatureFilter)
	for _, filter := range filters {
		featureFilters[filter.Name()] = filter
	}

	return &FeatureManager{
		featureProvider: provider,
		featureFilters:  featureFilters,
	}, nil
}

// IsEnabled determines if a feature flag is enabled.
// This is the primary method used to check feature flag state in application code.
//
// Parameters:
//   - featureName: The name of the feature to evaluate
//
// Returns:
//   - bool: true if the feature is enabled, false otherwise
//   - error: An error if the feature flag cannot be found or evaluated
func (fm *FeatureManager) IsEnabled(featureName string) (bool, error) {
	// Get the feature flag
	featureFlag, err := fm.featureProvider.GetFeatureFlag(featureName)
	if err != nil {
		return false, fmt.Errorf("failed to get feature flag %s: %w", featureName, err)
	}
	
	res, err := fm.evaluateFeature(featureFlag, nil)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate feature %s: %w", featureName, err)
	}

	return res.Enabled, nil
}

// IsEnabledWithAppContext determines if a feature flag is enabled for the given context.
// This version allows passing application-specific context for conditional evaluation.
//
// Parameters:
//   - featureName: The name of the feature to evaluate
//   - appContext: An optional context object for contextual evaluation
//
// Returns:
//   - bool: true if the feature is enabled, false otherwise
//   - error: An error if the feature flag cannot be found or evaluated
func (fm *FeatureManager) IsEnabledWithAppContext(featureName string, appContext any) (bool, error) {
	// Get the feature flag
	featureFlag, err := fm.featureProvider.GetFeatureFlag(featureName)
	if err != nil {
		return false, fmt.Errorf("failed to get feature flag %s: %w", featureName, err)
	}

	res, err := fm.evaluateFeature(featureFlag, appContext)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate feature %s: %w", featureName, err)
	}

	return res.Enabled, nil
}

// GetVariant returns the assigned variant for a feature flag.
// This method is used for implementing multivariate feature flags, A/B testing,
// or feature configurations that change based on the user context.
//
// Parameters:
//   - featureName: The name of the feature to evaluate
//   - appContext: An optional context object for contextual evaluation
//
// Returns:
//   - Variant: The assigned variant with its name and configuration value
//   - error: An error if the feature flag cannot be found or evaluated
func (fm *FeatureManager) GetVariant(featureName string, appContext any) (Variant, error) {
	// Implementation would be here
	return Variant{}, nil
}

// OnFeatureEvaluated registers a callback function that is invoked whenever a feature flag is evaluated.
// This method enables tracking feature usage, logging evaluation results, or implementing custom
// telemetry when features are checked.
//
// The registered callback receives an EvaluationResult struct containing details about the
// feature evaluation.
func (fm *FeatureManager) OnFeatureEvaluated(callback func(evalRes EvaluationResult)) {
	if callback == nil {
		return
	}

	fm.onFeatureEvaluated = append(fm.onFeatureEvaluated, callback)
}

// GetFeatureNames returns the names of all available features.
//
// Returns:
//   - []string: A slice containing the names of all available features
func (fm *FeatureManager) GetFeatureNames() []string {
	flags, err := fm.featureProvider.GetFeatureFlags()
	if err != nil {
		log.Printf("failed to get feature flag names: %v", err)
		return nil
	}

	res := make([]string, len(flags))
	for _, flag := range flags {
		res = append(res, flag.ID)
	}

	return res
}

func (fm *FeatureManager) isEnabled(featureFlag FeatureFlag, appContext any) (bool, error) {
	// If the feature is not explicitly enabled, then it is disabled by default
	if !featureFlag.Enabled {
		return false, nil
	}

	// If there are no client filters, then the feature is enabled
	if featureFlag.Conditions == nil || len(featureFlag.Conditions.ClientFilters) == 0 {
		return true, nil
	}

	// Default requirement type is "Any"
	requirementType := RequirementTypeAny
	if featureFlag.Conditions.RequirementType != "" {
		requirementType = featureFlag.Conditions.RequirementType
	}

	// Short circuit based on requirement type
	// - When "All", feature is enabled if all filters match (short circuit on false)
	// - When "Any", feature is enabled if any filter matches (short circuit on true)
	shortCircuitEvalResult := requirementType == RequirementTypeAny

	// Evaluate filters
	for _, clientFilter := range featureFlag.Conditions.ClientFilters {
		matchedFeatureFilter, exists := fm.featureFilters[clientFilter.Name]
		if !exists {
			log.Printf("Feature filter %s is not found", clientFilter.Name)
			return false, nil
		}

		// Create context with feature name and parameters
		filterContext := FeatureFilterEvaluationContext{
			FeatureName: featureFlag.ID,
			Parameters:  clientFilter.Parameters,
		}

		// Evaluate the filter
		filterResult, err := matchedFeatureFilter.Evaluate(filterContext, appContext)
		if err != nil {
			return false, fmt.Errorf("error evaluating filter %s: %w", clientFilter.Name, err)
		}

		// Short circuit if we hit the condition
		if filterResult == shortCircuitEvalResult {
			return shortCircuitEvalResult, nil
		}
	}

	// If we get here, we haven't short-circuited, so return opposite result
	return !shortCircuitEvalResult, nil
}

func (fm *FeatureManager) evaluateFeature(featureFlag FeatureFlag, appContext any) (EvaluationResult, error) {
	result := EvaluationResult{
		Feature: &featureFlag,
	}

	// Validate feature flag format
	if err := validateFeatureFlag(featureFlag); err != nil {
		return result, fmt.Errorf("invalid feature flag: %w", err)
	}

	// Evaluate if feature is enabled
	enabled, err := fm.isEnabled(featureFlag, appContext)
	if err != nil {
		return result, err
	}
	result.Enabled = enabled

	var targetingContext *TargetingContext
	if appContext != nil {
		if tc, ok := appContext.(*TargetingContext); ok {
			result.TargetingID = tc.UserID
			targetingContext = tc
		}
	}

	// Determine variant
	var variantDef *VariantDefinition
	reason := VariantAssignmentReasonNone

	// Process variants if present
	if len(featureFlag.Variants) > 0 {
		if !result.Enabled {
			reason = VariantAssignmentReasonDefaultWhenDisabled
			if featureFlag.Allocation != nil && featureFlag.Allocation.DefaultWhenDisabled != "" {
				variantDef = getVariant(featureFlag.Variants, featureFlag.Allocation.DefaultWhenDisabled)
			}
		} else {
			// Enabled, assign based on allocation
			if targetingContext != nil && featureFlag.Allocation != nil {
				if variantAssignment, err := assignVariant(featureFlag, *targetingContext); err == nil {
					variantDef = variantAssignment.Variant
					reason = variantAssignment.Reason
				}
			}

			// Allocation failed, assign default if specified
			if variantDef == nil && reason == VariantAssignmentReasonNone {
				reason = VariantAssignmentReasonDefaultWhenEnabled
				if featureFlag.Allocation != nil && featureFlag.Allocation.DefaultWhenEnabled != "" {
					variantDef = getVariant(featureFlag.Variants, featureFlag.Allocation.DefaultWhenEnabled)
				}
			}
		}
	}

	// Set variant in result
	if variantDef != nil {
		result.Variant = &Variant{
			Name:               variantDef.Name,
			ConfigurationValue: variantDef.ConfigurationValue,
		}
	}
	result.VariantAssignmentReason = reason

	// Apply status override from variant
	if variantDef != nil && featureFlag.Enabled {
		if variantDef.StatusOverride == StatusOverrideEnabled {
			result.Enabled = true
		} else if variantDef.StatusOverride == StatusOverrideDisabled {
			result.Enabled = false
		}
	}

	// Trigger callbacks if telemetry is enabled
	if featureFlag.Telemetry != nil && featureFlag.Telemetry.Enabled && len(fm.onFeatureEvaluated) > 0 {
		for _, callback := range fm.onFeatureEvaluated {
			if callback != nil {
				callback(result)
			}
		}
	}

	return result, nil
}

func getVariant(variants []VariantDefinition, name string) *VariantDefinition {
	for _, v := range variants {
		if v.Name == name {
			return &v
		}
	}

	return nil
}

type variantAssignment struct {
	Variant *VariantDefinition
	Reason  VariantAssignmentReason
}

func getVariantAssignment(featureFlag FeatureFlag, variantName string, reason VariantAssignmentReason) *variantAssignment {
	if variantName == "" {
		return nil
	}

	variant := getVariant(featureFlag.Variants, variantName)
	if variant == nil {
		log.Printf("Variant %s not found in feature %s", variantName, featureFlag.ID)
		return nil
	}

	return &variantAssignment{
		Variant: variant,
		Reason:  reason,
	}
}

func assignVariant(featureFlag FeatureFlag, targetingContext TargetingContext) (*variantAssignment, error) {
	if featureFlag.Allocation == nil {
		return nil, fmt.Errorf("no allocation defined for feature %s", featureFlag.ID)
	}

	if len(featureFlag.Allocation.User) > 0 {
		for _, userAlloc := range featureFlag.Allocation.User {
			if isTargetedUser(targetingContext.UserID, userAlloc.Users) {
				return getVariantAssignment(featureFlag, userAlloc.Variant, VariantAssignmentReasonUser), nil
			}
		}
	}

	if len(featureFlag.Allocation.Group) > 0 {
		for _, groupAlloc := range featureFlag.Allocation.Group {
			if isTargetedGroup(targetingContext.Groups, groupAlloc.Groups) {
				return getVariantAssignment(featureFlag, groupAlloc.Variant, VariantAssignmentReasonGroup), nil
			}
		}
	}

	if len(featureFlag.Allocation.Percentile) > 0 {
		for _, percentAlloc := range featureFlag.Allocation.Percentile {
			hint := featureFlag.Allocation.Seed
			if hint == "" {
				hint = fmt.Sprintf("allocation\n%s", featureFlag.ID)
			}

			if ok, _ := isTargetedPercentile(targetingContext.UserID, hint, percentAlloc.From, percentAlloc.To); ok {
				return getVariantAssignment(featureFlag, percentAlloc.Variant, VariantAssignmentReasonPercentile), nil
			}
		}
	}

	return &variantAssignment{
		Variant: nil,
		Reason:  VariantAssignmentReasonNone,
	}, nil
}
