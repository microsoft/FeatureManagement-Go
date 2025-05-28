package featuremanagement

// FeatureFlag represents a feature flag definition according to the v2.0.0 schema.
// Feature flags are used to dynamically enable or disable features in an application.
type FeatureFlag struct {
	// ID uniquely identifies the feature
	ID string `json:"id"`

	// Description provides details about the feature's purpose
	Description string `json:"description,omitempty"`

	// DisplayName is a human-friendly name for display purposes
	DisplayName string `json:"display_name,omitempty"`

	// Enabled indicates if the feature is on or off
	Enabled bool `json:"enabled"`

	// Conditions defines when the feature should be dynamically enabled
	Conditions *Conditions `json:"conditions,omitempty"`

	// Variants represents different configurations of this feature
	Variants []VariantDefinition `json:"variants,omitempty"`

	// Allocation determines how variants are assigned to users
	Allocation *VariantAllocation `json:"allocation,omitempty"`

	// Telemetry contains feature flag telemetry configuration
	Telemetry *Telemetry `json:"telemetry,omitempty"`
}

// Conditions defines the rules for enabling a feature dynamically.
type Conditions struct {
	// RequirementType determines if any or all filters must be satisfied.
	// Values: "Any" or "All"
	RequirementType string `json:"requirement_type,omitempty"`

	// ClientFilters are the filter conditions that must be evaluated by the client
	ClientFilters []ClientFilter `json:"client_filters,omitempty"`
}

// ClientFilter represents a filter that must be evaluated for feature enablement.
type ClientFilter struct {
	// Name is the identifier for this filter type
	Name string `json:"name"`

	// Parameters are the configuration values for the filter
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// Telemetry contains options for feature flag telemetry.
type Telemetry struct {
	// Enabled indicates if telemetry is enabled for this feature
	Enabled bool `json:"enabled,omitempty"`

	// Metadata contains additional data to include with telemetry
	Metadata map[string]string `json:"metadata,omitempty"`
}

// EvaluationResult contains information about a feature flag evaluation.
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
