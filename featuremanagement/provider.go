// Package featuremanagement provides functionality for feature flag management in Go applications.
// It allows developers to control feature releases and implement A/B testing through
// feature flags and variants.
package featuremanagement

// FeatureFlagProvider defines the interface for retrieving feature flags from a source.
// Implementations of this interface can fetch feature flags from various configuration
// stores such as Azure App Configuration, local JSON files, or other sources.
type FeatureFlagProvider interface {
	// GetFeatureFlag retrieves a specific feature flag by its name.
	//
	// Parameters:
	//   - name: The name of the feature flag to retrieve
	//
	// Returns:
	//   - FeatureFlag: The feature flag if found
	//   - error: An error if the feature flag cannot be found or retrieved
	GetFeatureFlag(name string) (FeatureFlag, error)

	// GetFeatureFlags retrieves all available feature flags.
	//
	// Returns:
	//   - []FeatureFlag: A slice of all available feature flags
	//   - error: An error if the feature flags cannot be retrieved
	GetFeatureFlags() ([]FeatureFlag, error)
}
