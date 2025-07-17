// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package featuremanagement

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/go-viper/mapstructure/v2"
)

type TargetingFilter struct{}

// TargetingGroup defines a named group with a specific rollout percentage
type TargetingGroup struct {
	Name              string
	RolloutPercentage float64
}

// TargetingExclusion defines users and groups explicitly excluded from targeting
type TargetingExclusion struct {
	Users  []string
	Groups []string
}

// TargetingAudience defines the targeting configuration for feature rollout
type TargetingAudience struct {
	DefaultRolloutPercentage float64
	Users                    []string
	Groups                   []TargetingGroup
	Exclusion                *TargetingExclusion
}

// TargetingFilterParameters defines the parameters for the targeting filter
type TargetingFilterParameters struct {
	Audience TargetingAudience
}

func (t *TargetingFilter) Name() string {
	return "Microsoft.Targeting"
}

func (t *TargetingFilter) Evaluate(evalCtx FeatureFilterEvaluationContext, appCtx any) (bool, error) {
	// Validate parameters
	params, err := getTargetingParams(evalCtx)
	if err != nil {
		return false, err
	}

	// Check if app context is valid
	targetingCtx, ok := appCtx.(TargetingContext)
	if !ok {
		return false, fmt.Errorf("the app context is required for targeting filter and must be of type TargetingContext")
	}

	// Check exclusions
	if params.Audience.Exclusion != nil {
		// Check if the user is in the exclusion list
		if targetingCtx.UserID != "" &&
			len(params.Audience.Exclusion.Users) > 0 &&
			isTargetedUser(targetingCtx.UserID, params.Audience.Exclusion.Users) {
			return false, nil
		}

		// Check if the user is in a group within exclusion list
		if len(targetingCtx.Groups) > 0 &&
			len(params.Audience.Exclusion.Groups) > 0 &&
			isTargetedGroup(targetingCtx.Groups, params.Audience.Exclusion.Groups) {
			return false, nil
		}
	}

	// Check if the user is being targeted directly
	if targetingCtx.UserID != "" &&
		len(params.Audience.Users) > 0 &&
		isTargetedUser(targetingCtx.UserID, params.Audience.Users) {
		return true, nil
	}

	// Check if the user is in a group that is being targeted
	if len(targetingCtx.Groups) > 0 && len(params.Audience.Groups) > 0 {
		for _, group := range params.Audience.Groups {
			if isTargetedGroup(targetingCtx.Groups, []string{group.Name}) {
				// Check if user is in the rollout percentage for this group
				hint := fmt.Sprintf("%s\n%s", evalCtx.FeatureName, group.Name)
				targeted, err := isTargetedPercentile(targetingCtx.UserID, hint, 0, group.RolloutPercentage)
				if err != nil {
					return false, err
				}
				if targeted {
					return true, nil
				}
			}
		}
	}

	// Check if the user is being targeted by a default rollout percentage
	hint := evalCtx.FeatureName
	return isTargetedPercentile(targetingCtx.UserID, hint, 0, params.Audience.DefaultRolloutPercentage)
}

func getTargetingParams(evalCtx FeatureFilterEvaluationContext) (TargetingFilterParameters, error) {
	var params TargetingFilterParameters
	err := mapstructure.Decode(evalCtx.Parameters, &params)
	if err != nil {
		return TargetingFilterParameters{}, fmt.Errorf("failed to decode feature flag parameters: %v", err)
	}

	// Validate DefaultRolloutPercentage
	if params.Audience.DefaultRolloutPercentage < 0 || params.Audience.DefaultRolloutPercentage > 100 {
		return TargetingFilterParameters{}, fmt.Errorf("invalid feature flag: %s. Audience.DefaultRolloutPercentage must be a number between 0 and 100", evalCtx.FeatureName)
	}

	// Validate RolloutPercentage for each group
	if len(params.Audience.Groups) > 0 {
		for _, group := range params.Audience.Groups {
			if group.RolloutPercentage < 0 || group.RolloutPercentage > 100 {
				return TargetingFilterParameters{}, fmt.Errorf("invalid feature flag: %s. RolloutPercentage of group %s must be a number between 0 and 100", evalCtx.FeatureName, group.Name)
			}
		}
	}

	return params, nil
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
