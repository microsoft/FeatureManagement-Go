// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package featuremanagement

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type TimeWindowFilter struct{}

type TimeWindowFilterParameters struct {
	Start string `json:"start,omitempty"`
	End   string `json:"end,omitempty"`
}

func (t *TimeWindowFilter) Name() string {
	return "Microsoft.TimeWindow"
}

func (t *TimeWindowFilter) Evaluate(evalCtx FeatureFilterEvaluationContext, appContext any) (bool, error) {
	// Extract and parse parameters
	paramsBytes, err := json.Marshal(evalCtx.Parameters)
	if err != nil {
		return false, fmt.Errorf("failed to marshal time window parameters: %w", err)
	}

	var params TimeWindowFilterParameters
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		return false, fmt.Errorf("invalid time window parameters format: %w", err)
	}

	var startTime, endTime *time.Time

	// Parse start time if provided
	if params.Start != "" {
		parsed, err := time.Parse(time.RFC3339, params.Start)
		if err != nil {
			return false, fmt.Errorf("invalid start time format for feature %s: %w", evalCtx.FeatureName, err)
		}
		startTime = &parsed
	}

	// Parse end time if provided
	if params.End != "" {
		parsed, err := time.Parse(time.RFC3339, params.End)
		if err != nil {
			return false, fmt.Errorf("invalid end time format for feature %s: %w", evalCtx.FeatureName, err)
		}
		endTime = &parsed
	}

	// Check if at least one time parameter exists
	if startTime == nil && endTime == nil {
		log.Printf("The Microsoft.TimeWindow feature filter is not valid for feature %s. It must specify either 'Start', 'End', or both.", evalCtx.FeatureName)
		return false, nil
	}

	// Get current time
	now := time.Now()

	// Check if current time is within the window
	// (after or equal to start time AND before end time)
	isAfterStart := startTime == nil || !now.Before(*startTime)
	isBeforeEnd := endTime == nil || now.Before(*endTime)

	return isAfterStart && isBeforeEnd, nil
}
