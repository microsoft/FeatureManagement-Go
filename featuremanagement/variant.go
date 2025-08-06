// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package featuremanagement

// Variant represents a feature configuration variant.
// Variants allow different configurations or implementations of a feature
// to be assigned to different users.
type Variant struct {
	// Name uniquely identifies this variant
	Name string

	// ConfigurationValue holds the value for this variant
	ConfigurationValue any
}
