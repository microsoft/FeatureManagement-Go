// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package featuremanagement

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
)

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

// isTargetedPercentile determines if the user is part of the audience based on percentile range
func isTargetedPercentile(userID string, hint string, from float64, to float64) (bool, error) {
	// Validate percentile range
	if from < 0 || from > 100 {
		return false, fmt.Errorf("the 'from' value must be between 0 and 100")
	}
	if to < 0 || to > 100 {
		return false, fmt.Errorf("the 'to' value must be between 0 and 100")
	}
	if from > to {
		return false, fmt.Errorf("the 'from' value cannot be larger than the 'to' value")
	}

	audienceContextID := constructAudienceContextID(userID, hint)

	// Convert to uint32 for percentage calculation
	contextMarker, err := hashStringToUint32(audienceContextID)
	if err != nil {
		return false, err
	}

	// Calculate percentage (0-100)
	contextPercentage := (float64(contextMarker) / float64(math.MaxUint32)) * 100

	// Handle edge case of exact 100 bucket
	if to == 100 {
		return contextPercentage >= from, nil
	}

	return contextPercentage >= from && contextPercentage < to, nil
}

// isTargetedGroup determines if the user is part of the audience based on groups
func isTargetedGroup(sourceGroups []string, targetedGroups []string) bool {
	if len(sourceGroups) == 0 {
		return false
	}

	// Check if any source group is in the targeted groups
	for _, sourceGroup := range sourceGroups {
		for _, targetedGroup := range targetedGroups {
			if sourceGroup == targetedGroup {
				return true
			}
		}
	}

	return false
}

// isTargetedUser determines if the user is part of the audience based on user ID
func isTargetedUser(userID string, users []string) bool {
	if userID == "" {
		return false
	}

	// Check if the user is in the targeted users list
	for _, user := range users {
		if userID == user {
			return true
		}
	}

	return false
}

// constructAudienceContextID builds the context ID for the audience
func constructAudienceContextID(userID string, hint string) string {
	return fmt.Sprintf("%s\n%s", userID, hint)
}

// hashStringToUint32 converts a string to a uint32 using SHA-256 hashing
func hashStringToUint32(s string) (uint32, error) {
	hash := sha256.Sum256([]byte(s))
	// Extract first 4 bytes and convert to uint32 (little-endian)
	return binary.LittleEndian.Uint32(hash[:4]), nil
}
