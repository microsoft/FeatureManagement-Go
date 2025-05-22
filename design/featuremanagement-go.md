<p>Feature Management - Go 
Overview 
Developers can use feature flags in simple use cases like conditional statement or more advanced scenarios in Golang.</p>
<p>Design Proposal 
Feature management will be implemented as a new package <code>featuremanagement</code>. 
This package will:
Support loading feature flags from various sources, not limited to Azure App Configuration
Evaluate Feature Flags with and without filters 
Default filters
Custom filters
Requirement Type All and Any 
Support Variants
Provide telemetry hooks</p>
<p>API</p>
<p>Package name: featuremanagement
Github repo: github.com/microsoft/Featuremanagement-Go
Import path: github.com/microsoft/Featuremanagement-Go/featuremanagement</p>
<p>// In Go, if users want to use an exported function from a package, they need to:
// 1. Import the package
//  Use the import statement to bring the package into the file.
//
// 2. Call the exported function
//  Use the package name followed by a dot (.) and the function name.
//  Note: In Go, exported functions (or variables/types) start with a capital letter.
// import (
//     // package location, installed via <code>go get</code>
//     "github.com/microsoft/Featuremanagement-Go/featuremanagement" New
// )</p>
<p>type FeatureFlagProvider interface {
    GetFeatureFlag(name string) (FeatureFlag, error)
    GetFeatureFlags() ([]FeatureFlag, error)
}</p>
<p>// AzureAppConfigurationFeatureFlagProvider implements the FeatureFlagProvider 
// interface for Azure App Configuration. It retrieves feature flags directly 
// from an Azure App Configuration instance.
//
//
// Usage:
//
//   appConfig, _ := azureappconfiguration.Load(context.Background(), authOptions, configOptions)
//   
//   provider := &amp;AzureAppConfigurationFeatureFlagProvider{
//     AzureAppConfiguration: appConfig,
//   }
//
//   // Use the provider to initialize a feature manager
//   manager, _ := featuremanagement.NewFeatureManager(provider, nil)
//</p>
<p>Sub-package
type AzureAppConfigurationFeatureFlagProvider struct {
    AzureAppConfiguration *azureappconfiguation.AzureAppConfiguration
}</p>
<p>type BytesFeatureFlagProvider struct {  =&gt; optional
  Bytes []bytes
}</p>
<p>// NewFeatureManager creates and initializes a new instance of the FeatureManager.
// This is the entry point for using feature management functionality.
//
// Parameters:
//   - provider: A FeatureFlagProvider that supplies feature flag definitions
//     from a source such as Azure App Configuration or a local JSON file
//   - filters: Custom filters for conditional feature evaluation
//   
// Returns:
//   - <em>FeatureManager: A configured feature manager instance ready for use
//   - error: An error if initialization fails
//
func NewFeatureManager(provider FeatureFlagProvider, filters []FeatureFilter) (</em>FeatureManager, error)</p>
<p>// IsEnabled determines if a feature flag is enabled for the given context.
// This is the primary method used to check feature flag state in application code.
//
// Parameters:
//   - featureName: The name of the feature to evaluate
//   - appContext: An optional context object for contextual evaluation. This can be:
//     * nil - for simple feature checks without user context
//     * TargetingContext - for user-specific targeting
//     * A custom context type used by custom filters
//
// Returns:
//   - bool: true if the feature is enabled, false otherwise
//   - error: An error if the feature flag cannot be found or evaluated
//
// Usage Examples:
//
//   // Simple feature check (no user context)
//   if enabled, _ := manager.IsEnabled("dark-mode", nil); enabled {
//       ui.SetTheme(DarkTheme)
//   }
//
//   // User-specific feature check
//   ctx := &amp;featuremanagement.TargetingContext{
//       UserId: "user-12345",
//       Groups: []string{"beta-testers"},
//   }
//
//   if enabled, err := manager.IsEnabled("new-ui", ctx); err != nil {
//       log.Printf("Error evaluating feature: %v", err)
//       return fallbackUI() // Handle error case
//   } else if enabled {
//       return newUI()
//   } else {
//       return oldUI()
//   }
func (fm *FeatureManager) IsEnabled(featureName string) (bool, error)
IsEnabled(featureName string)
IsEnabledWithAppContext(featureName string, appContext any)</p>
<p>type TargetingContext struct {
    // User identifier for targeting specific users
    UserId string
    
    // Groups the user belongs to for group targeting
    Groups []string
}</p>
<p>// GetVariant returns the assigned variant for a feature flag.
// This method is used for implementing multivariate feature flags, A/B testing,
// or feature configurations that change based on the user context.
//
// Parameters:
//   - featureName: The name of the feature to evaluate
//   - appContext: An optional context object for contextual evaluation. This can be:
//     * nil - for simple feature checks without user context
//     * TargetingContext - for user-specific targeting and variant assignment
//     * A custom context type used by custom filters
//
// Returns:
//   - Variant: The assigned variant with its name and configuration value
//   - error: An error if the feature flag cannot be found or evaluated
//
// Usage Example:
//
//   ctx := &amp;featuremanagement.TargetingContext{
//       UserId: "user-12345",
//       Groups: []string{"premium-tier"},
//   }
//
//   variant, err := manager.GetVariant("pricing-model", ctx)
//   if err != nil {
//       // Handle error (feature disabled or not found)
//       return defaultPricingModel()
//   }
//
//   switch variant.Name {
//   case "standard":
//       return standardPricing()
//   case "discount":
//       discount := 10.0 // Default
//       if val, ok := variant.ConfigurationValue.(float64); ok {
//           discount = val
//       }
//       return discountPricing(discount)
//   default:
//       return defaultPricingModel()
//   }
func (fm *FeatureManager) GetVariant(featureName string, appContext any) (Variant, error)</p>
<p>type Variant struct {
    Name string <code>json:"name"</code>
    // ConfigurationValue is the value for this variant
    ConfigurationValue any <code>json:"configuration_value,omitempty"</code>
}</p>
<p>type FeatureFilterEvaluationContext struct {
    FeatureName string
    Parameters map[string]any
}</p>
<p>// FeatureFilter defines the interface for feature flag filters.
//
// Example custom filter:
//
//   type EnvironmentFilter struct{}
//
//   func (f EnvironmentFilter) Name() string {
//       return "EnvironmentFilter"
//   }
//
//   func (f EnvironmentFilter) (evalCtx FeatureFilterEvaluationContext, appCtx any) (bool, error) {
//       // Implementation
//       // ...
//   }
//
//   // Register custom filter with feature manager
//   manager, _ := featuremanagement.NewFeatureManager(
//       provider,
//       []featuremanagement.FeatureFilter{&amp;EnvironmentFilter{}},
//   )
//
type FeatureFilter interface {
    Name() string
    Evaluate(evalCtx FeatureFilterEvaluationContext, appCtx any) (bool, error)
}</p>
<p>// EvaluationResult contains details about feature evaluation
type EvaluationResult struct {
    Feature *FeatureFlag</p>
<p>// Enabled state
    Enabled bool</p>
<p>// Variant assignment
    TargetingID            string
    Variant                *Variant
    VariantAssignmentReason VariantAssignmentReason
}</p>
<p>// OnFeatureEvaluated registers a callback function that is invoked whenever a feature flag is evaluated.
// This method enables tracking feature usage, logging evaluation results, or implementing custom
// telemetry when features are checked.
//
// The registered callback receives an EvaluationResult struct containing details about the 
// feature evaluation, including:
// - The feature flag details
// - Whether the feature was enabled
// - The targeting ID used for evaluation
// - Any variant assignment information
//
// Usage Example:
//
//   // Register a callback for metrics collection
//   manager.OnFeatureEvaluated(func(evalRes EvaluationResult) {
//       metrics.RecordFeatureEvaluation(
//           evalRes.Feature.ID,
//           evalRes.Enabled,
//           evalRes.Variant != nil,
//       )
//   })
//
func (fm *FeatureManager) OnFeatureEvaluated(func(evalRes EvaluationResult))</p>
<p>func (fm *FeatureManger) GetFeatureNames() []string</p>
<p>Feature Flag Schema</p>
<p>Feature Flag related struct definition is based on FeatureManagement/Schema/FeatureFlag.v2.0.0.schema.json at main · microsoft/FeatureManagement</p>
<p>// FeatureFlag represents a feature flag definition according to the v2.0.0 schema
type FeatureFlag struct {
    // ID uniquely identifies the feature
    ID string <code>json:"id"</code>
    // Description provides details about the feature's purpose
    Description string <code>json:"description,omitempty"</code>
    // DisplayName is a human-friendly name for display purposes
    DisplayName string <code>json:"display_name,omitempty"</code>
    // Enabled indicates if the feature is on or off
    Enabled bool <code>json:"enabled"</code>
    // Conditions defines when the feature should be dynamically enabled
    Conditions <em>Conditions <code>json:"conditions,omitempty"</code>
    // Variants represents different configurations of this feature
    Variants []Variant <code>json:"variants,omitempty"</code>
    // Allocation determines how variants are assigned to users
    Allocation </em>VariantAllocation <code>json:"allocation,omitempty"</code>
    // Telemetry contains feature flag telemetry configuration
    Telemetry *Telemetry <code>json:"telemetry,omitempty"</code>
}</p>
<p>// Conditions defines the rules for enabling a feature dynamically
type Conditions struct {
    // RequirementType determines if any or all filters must be satisfied
    // Values: "Any" or "All"
    RequirementType string <code>json:"requirement_type,omitempty"</code>
    // ClientFilters are the filter conditions that must be evaluated by the client
    ClientFilters []ClientFilter <code>json:"client_filters,omitempty"</code>
}</p>
<p>// ClientFilter represents a filter that must be evaluated for feature enablement
type ClientFilter struct {
    // Name is the identifier for this filter type
    Name string <code>json:"name"</code>
    // Parameters are the configuration values for the filter
    Parameters map[string]interface{} <code>json:"parameters,omitempty"</code>
}</p>
<p>// Variant represents a feature configuration variant
type Variant struct {
    // Name uniquely identifies this variant
    Name string <code>json:"name"</code>
    // ConfigurationValue holds the value for this variant 
    ConfigurationValue interface{} <code>json:"configuration_value,omitempty"</code>
    // StatusOverride overrides the enabled state of the feature when this variant is assigned
    // Values: "None", "Enabled", "Disabled"
    StatusOverride string <code>json:"status_override,omitempty"</code>
}</p>
<p>// VariantAllocation defines rules for assigning variants to users
type VariantAllocation struct {
    // DefaultWhenDisabled specifies which variant to use when feature is disabled
    DefaultWhenDisabled string <code>json:"default_when_disabled,omitempty"</code>
    // DefaultWhenEnabled specifies which variant to use when feature is enabled
    DefaultWhenEnabled string <code>json:"default_when_enabled,omitempty"</code>
    // User defines variant assignments for specific users
    User []UserAllocation <code>json:"user,omitempty"</code>
    // Group defines variant assignments for user groups
    Group []GroupAllocation <code>json:"group,omitempty"</code>
    // Percentile defines variant assignments by percentage ranges
    Percentile []PercentileAllocation <code>json:"percentile,omitempty"</code>
    // Seed is used to ensure consistent percentile calculations across features
    Seed string <code>json:"seed,omitempty"</code>
}</p>
<p>// UserAllocation assigns a variant to specific users
type UserAllocation struct {
    // Variant is the name of the variant to use
    Variant string <code>json:"variant"</code>
    // Users is the collection of user IDs to apply this variant to
    Users []string <code>json:"users"</code>
}</p>
<p>// GroupAllocation assigns a variant to specific user groups
type GroupAllocation struct {
    // Variant is the name of the variant to use
    Variant string <code>json:"variant"</code>
    // Groups is the collection of group IDs to apply this variant to
    Groups []string <code>json:"groups"</code>
}</p>
<p>// PercentileAllocation assigns a variant to a percentage range of users
type PercentileAllocation struct {
    // Variant is the name of the variant to use
    Variant string <code>json:"variant"</code>
    // From is the lower end of the percentage range (0-100)
    From float64 <code>json:"from"</code>
    // To is the upper end of the percentage range (0-100)
    To float64 <code>json:"to"</code>
}</p>
<p>// Telemetry contains options for feature flag telemetry
type Telemetry struct {
    // Enabled indicates if telemetry is enabled for this feature
    Enabled bool <code>json:"enabled,omitempty"</code>
    // Metadata contains additional data to include with telemetry
    Metadata map[string]string <code>json:"metadata,omitempty"</code>
}</p>
<p>// VariantAssignmentReason represents the reason a variant was assigned
type VariantAssignmentReason string</p>
<p>const (
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
)</p>
<p>// EvaluationResult contains information about a feature flag evaluation
type EvaluationResult struct {
    // Feature contains the evaluated feature flag
    Feature <em>FeatureFlag
    // Enabled indicates the final state of the feature after evaluation
    Enabled bool
    // TargetingID is the identifier used for consistent targeting
    TargetingID string
    // Variant is the selected variant (if any)
    Variant </em>Variant
    // VariantAssignmentReason explains why the variant was assigned
    VariantAssignmentReason VariantAssignmentReason
}</p>
<p>Updates after design review meeting 05/14:
IsEnabled(featureName string)
IsEnabledWithAppContext(featureName string, appContext any)</p>
<p>Sub-module for AzureAppConfigurationFeatureFlagProvider</p>
<p>BytesFeatureFlagProvider : optional</p>
<p>Discussions 
Any open questions for further discussion? Anything else will be impacted or should be involved? 
Revision History 
 </p>