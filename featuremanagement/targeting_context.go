package featuremanagement

// TargetingContext provides user-specific information for feature flag targeting.
// This is used to determine if a feature should be enabled for a specific user
// or to select the appropriate variant for a user.
type TargetingContext struct {
	// UserID is the identifier for targeting specific users
	UserID string

	// Groups are the groups the user belongs to for group targeting
	Groups []string
}
