// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package featuremanagement

import (
	"encoding/json"
	"testing"
)

func TestGetVariant(t *testing.T) {
	jsonData := `{
		"feature_flags": [
            {
                "id": "VariantFeaturePercentileOn",
                "enabled": true,
                "variants": [
                    {
                        "name": "Big",
                        "status_override": "Disabled"
                    }
                ],
                "allocation": {
                    "percentile": [
                        {
                            "variant": "Big",
                            "from": 0,
                            "to": 50
                        }
                    ],
                    "seed": "1234"
                },
                "telemetry": {
                    "enabled": true
                }
            },
            {
                "id": "VariantFeaturePercentileOff",
                "enabled": true,
                "variants": [
                    {
                        "name": "Big"
                    }
                ],
                "allocation": {
                    "percentile": [
                        {
                            "variant": "Big",
                            "from": 0,
                            "to": 50
                        }
                    ],
                    "seed": "12345"
                },
                "telemetry": {
                    "enabled": true
                }
            },
            {
                "id": "VariantFeatureDefaultDisabled",
                "enabled": false,
                "variants": [
                    {
                        "name": "Small",
                        "configuration_value": "300px"
                    }
                ],
                "allocation": {
                    "default_when_disabled": "Small"
                },
                "telemetry": {
                    "enabled": true
                }
            },
            {
                "id": "VariantFeatureDefaultEnabled",
                "enabled": true,
                "variants": [
                    {
                        "name": "Medium",
                        "configuration_value": {
                            "Size": "450px",
                            "Color": "Purple"
                        }
                    },
                    {
                        "name": "Small",
                        "configuration_value": "300px"
                    }
                ],
                "allocation": {
                    "default_when_enabled": "Medium",
                    "user": [
                        {
                            "variant": "Small",
                            "users": [
                                "Jeff"
                            ]
                        }
                    ]
                },
                "telemetry": {
                    "enabled": true
                }
            },
            {
                "id": "VariantFeatureUser",
                "enabled": true,
                "variants": [
                    {
                        "name": "Small",
                        "configuration_value": "300px"
                    }
                ],
                "allocation": {
                    "user": [
                        {
                            "variant": "Small",
                            "users": [
                                "Marsha"
                            ]
                        }
                    ]
                },
                "telemetry": {
                    "enabled": true
                }
            },
            {
                "id": "VariantFeatureGroup",
                "enabled": true,
                "variants": [
                    {
                        "name": "Small",
                        "configuration_value": "300px"
                    }
                ],
                "allocation": {
                    "group": [
                        {
                            "variant": "Small",
                            "groups": [
                                "Group1"
                            ]
                        }
                    ]
                },
                "telemetry": {
                    "enabled": true
                }
            },
            {
                "id": "VariantFeatureNoVariants",
                "enabled": true,
                "variants": [],
                "allocation": {
                    "user": [
                        {
                            "variant": "Small",
                            "users": [
                                "Marsha"
                            ]
                        }
                    ]
                },
                "telemetry": {
                    "enabled": true
                }
            },
            {
                "id": "VariantFeatureNoAllocation",
                "enabled": true,
                "variants": [
                    {
                        "name": "Small",
                        "configuration_value": "300px"
                    }
                ],
                "telemetry": {
                    "enabled": true
                }
            }
        ]
    }`

	// Parse the feature flags configuration
	var featureManagement struct {
		FeatureFlags []FeatureFlag `json:"feature_flags"`
	}

	if err := json.Unmarshal([]byte(jsonData), &featureManagement); err != nil {
		t.Fatalf("Failed to unmarshal feature flags: %v", err)
	}

	// Create mock provider with the parsed feature flags
	provider := &mockFeatureFlagProvider{
		featureFlags: featureManagement.FeatureFlags,
	}

	// Create feature manager
	manager, err := NewFeatureManager(provider, nil)
	if err != nil {
		t.Fatalf("Failed to create feature manager: %v", err)
	}

	// Common test context
	context := TargetingContext{
		UserID: "Marsha",
		Groups: []string{"Group1"},
	}

	// Test valid scenarios
	t.Run("Valid scenarios", func(t *testing.T) {
		t.Run("Default allocation with disabled feature", func(t *testing.T) {
			variant, err := manager.GetVariant("VariantFeatureDefaultDisabled", context)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if variant.Name != "Small" {
				t.Errorf("Expected variant name 'Small', got '%s'", variant.Name)
			}

			if variant.ConfigurationValue != "300px" {
				t.Errorf("Expected configuration value '300px', got '%v'", variant.ConfigurationValue)
			}
		})

		t.Run("Default allocation with enabled feature", func(t *testing.T) {
			variant, err := manager.GetVariant("VariantFeatureDefaultEnabled", context)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if variant.Name != "Medium" {
				t.Errorf("Expected variant name 'Medium', got '%s'", variant.Name)
			}

			// Assert configuration value as map
			configMap, ok := variant.ConfigurationValue.(map[string]interface{})
			if !ok {
				t.Errorf("Expected configuration value to be a map, got %T", variant.ConfigurationValue)
			} else {
				size, sizeOk := configMap["Size"].(string)
				color, colorOk := configMap["Color"].(string)

				if !sizeOk || size != "450px" {
					t.Errorf("Expected Size '450px', got '%v'", size)
				}

				if !colorOk || color != "Purple" {
					t.Errorf("Expected Color 'Purple', got '%v'", color)
				}
			}
		})

		t.Run("User allocation", func(t *testing.T) {
			variant, err := manager.GetVariant("VariantFeatureUser", context)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if variant.Name != "Small" {
				t.Errorf("Expected variant name 'Small', got '%s'", variant.Name)
			}

			if variant.ConfigurationValue != "300px" {
				t.Errorf("Expected configuration value '300px', got '%v'", variant.ConfigurationValue)
			}
		})

		t.Run("Group allocation", func(t *testing.T) {
			variant, err := manager.GetVariant("VariantFeatureGroup", context)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if variant.Name != "Small" {
				t.Errorf("Expected variant name 'Small', got '%s'", variant.Name)
			}

			if variant.ConfigurationValue != "300px" {
				t.Errorf("Expected configuration value '300px', got '%v'", variant.ConfigurationValue)
			}
		})

		t.Run("Percentile allocation with seed", func(t *testing.T) {
			// First variant should be defined
			variant, err := manager.GetVariant("VariantFeaturePercentileOn", context)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if variant.Name != "Big" {
				t.Errorf("Expected variant name 'Big', got '%s'", variant.Name)
			}

			// Second variant should be undefined due to different seed
			_, err = manager.GetVariant("VariantFeaturePercentileOff", context)
			if err == nil {
				t.Error("Expected error for undefined variant, but got none")
			}
		})

		t.Run("Status override affecting enabled status", func(t *testing.T) {
			// The variant has status_override: "Disabled"
			enabled, err := manager.IsEnabledWithAppContext("VariantFeaturePercentileOn", context)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if enabled {
				t.Error("Expected feature to be disabled due to variant status override, but it's enabled")
			}
		})
	})
}
