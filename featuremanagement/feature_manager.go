// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package featuremanagement

import (
	"fmt"
	"log"
)

// FeatureManager is responsible for evaluating feature flags and their variants.
// It is the main entry point for interacting with the feature management library.
type FeatureManager struct {
	featureProvider FeatureFlagProvider
	featureFilters  map[string]FeatureFilter
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
		if filter != nil {
			featureFilters[filter.Name()] = filter
		}
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

	res, err := fm.isEnabled(featureFlag, nil)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate feature %s: %w", featureName, err)
	}

	return res, nil
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

	res, err := fm.isEnabled(featureFlag, appContext)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate feature %s: %w", featureName, err)
	}

	return res, nil
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

	res := make([]string, 0, len(flags))
	for i, flag := range flags {
		res[i] = flag.ID
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
