package featuremanagement

// VariantDefinition represents the configuration definition of a feature variant.
// It defines the properties of a variant in the feature flag configuration.
type VariantDefinition struct {
	// Name uniquely identifies this variant
	Name string `json:"name"`

	// ConfigurationValue holds the value for this variant
	ConfigurationValue interface{} `json:"configuration_value,omitempty"`

	// StatusOverride overrides the enabled state of the feature when this variant is assigned
	// Values: "None", "Enabled", "Disabled"
	StatusOverride string `json:"status_override,omitempty"`
}

// Variant represents a feature configuration variant.
// Variants allow different configurations or implementations of a feature
// to be assigned to different users.
type Variant struct {
	// Name uniquely identifies this variant
	Name string `json:"name"`

	// ConfigurationValue holds the value for this variant
	ConfigurationValue interface{} `json:"configuration_value,omitempty"`
}

// VariantAllocation defines rules for assigning variants to users.
type VariantAllocation struct {
	// DefaultWhenDisabled specifies which variant to use when feature is disabled
	DefaultWhenDisabled string `json:"default_when_disabled,omitempty"`

	// DefaultWhenEnabled specifies which variant to use when feature is enabled
	DefaultWhenEnabled string `json:"default_when_enabled,omitempty"`

	// User defines variant assignments for specific users
	User []UserAllocation `json:"user,omitempty"`

	// Group defines variant assignments for user groups
	Group []GroupAllocation `json:"group,omitempty"`

	// Percentile defines variant assignments by percentage ranges
	Percentile []PercentileAllocation `json:"percentile,omitempty"`

	// Seed is used to ensure consistent percentile calculations across features
	Seed string `json:"seed,omitempty"`
}

// UserAllocation assigns a variant to specific users.
type UserAllocation struct {
	// Variant is the name of the variant to use
	Variant string `json:"variant"`

	// Users is the collection of user IDs to apply this variant to
	Users []string `json:"users"`
}

// GroupAllocation assigns a variant to specific user groups.
type GroupAllocation struct {
	// Variant is the name of the variant to use
	Variant string `json:"variant"`

	// Groups is the collection of group IDs to apply this variant to
	Groups []string `json:"groups"`
}

// PercentileAllocation assigns a variant to a percentage range of users.
type PercentileAllocation struct {
	// Variant is the name of the variant to use
	Variant string `json:"variant"`

	// From is the lower end of the percentage range (0-100)
	From float64 `json:"from"`

	// To is the upper end of the percentage range (0-100)
	To float64 `json:"to"`
}

// VariantAssignmentReason represents the reason a variant was assigned.
type VariantAssignmentReason string

const (
	// VariantAssignmentReasonNone indicates no specific reason for variant assignment
	VariantAssignmentReasonNone VariantAssignmentReason = "None"

	// VariantAssignmentReasonDefaultWhenDisabled indicates the variant was assigned because it's the default for disabled features
	VariantAssignmentReasonDefaultWhenDisabled VariantAssignmentReason = "DefaultWhenDisabled"

	// VariantAssignmentReasonDefaultWhenEnabled indicates the variant was assigned because it's the default for enabled features
	VariantAssignmentReasonDefaultWhenEnabled VariantAssignmentReason = "DefaultWhenEnabled"

	// VariantAssignmentReasonUser indicates the variant was assigned based on the user's ID
	VariantAssignmentReasonUser VariantAssignmentReason = "User"

	// VariantAssignmentReasonGroup indicates the variant was assigned based on the user's group
	VariantAssignmentReasonGroup VariantAssignmentReason = "Group"

	// VariantAssignmentReasonPercentile indicates the variant was assigned based on percentile calculations
	VariantAssignmentReasonPercentile VariantAssignmentReason = "Percentile"
)
