// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package featuremanagement

// FeatureFilterEvaluationContext provides the context information needed
// to evaluate a feature filter.
type FeatureFilterEvaluationContext struct {
	// FeatureName is the name of the feature being evaluated
	FeatureName string

	// Parameters contains the filter-specific configuration parameters
	Parameters map[string]any
}

// TargetingContext provides user-specific information for feature flag targeting.
// This is used to determine if a feature should be enabled for a specific user
// or to select the appropriate variant for a user.
type TargetingContext struct {
	// UserID is the identifier for targeting specific users
	UserID string

	// Groups are the groups the user belongs to for group targeting
	Groups []string
}

// FeatureFilter defines the interface for feature flag filters.
// Filters determine whether a feature should be enabled based on certain conditions.
//
// Example custom filter:
//
//	type EnvironmentFilter struct{}
//
//	func (f EnvironmentFilter) Name() string {
//		return "EnvironmentFilter"
//	}
//
//	func (f EnvironmentFilter) Evaluate(evalCtx FeatureFilterEvaluationContext, appCtx any) (bool, error) {
//		// Implementation
//		// ...
//	}
//
//	// Register custom filter with feature manager
//	manager, _ := featuremanagement.NewFeatureManager(
//		provider,
//		[]featuremanagement.FeatureFilter{&EnvironmentFilter{}},
//	)
type FeatureFilter interface {
	// Name returns the identifier for this filter
	Name() string

	// Evaluate determines whether a feature should be enabled based on the provided contexts
	Evaluate(evalCtx FeatureFilterEvaluationContext, appCtx any) (bool, error)
}
