// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Azure/AppConfiguration-GoProvider/azureappconfiguration"
	"github.com/microsoft/Featuremanagement-Go/featuremanagement"
	"github.com/microsoft/Featuremanagement-Go/featuremanagement/providers/azappconfig"
)

func loadAzureAppConfiguration(ctx context.Context) (*azureappconfiguration.AzureAppConfiguration, error) {
	// Get the connection string from environment variable
	connectionString := os.Getenv("AZURE_APPCONFIG_CONNECTION_STRING")
	if connectionString == "" {
		return nil, fmt.Errorf("AZURE_APPCONFIG_CONNECTION_STRING environment variable is not set")
	}

	// Set up authentication options with connection string
	authOptions := azureappconfiguration.AuthenticationOptions{
		ConnectionString: connectionString,
	}

	// Configure which keys to load and trimming options
	options := &azureappconfiguration.Options{
		// Enable feature flags
		FeatureFlagOptions: azureappconfiguration.FeatureFlagOptions{
			Enabled: true,
			Selectors: []azureappconfiguration.Selector{
				{
					KeyFilter:   "*", // Load all feature flags
					LabelFilter: "",
				},
			},
			RefreshOptions: azureappconfiguration.RefreshOptions{
				Enabled: true,
			},
		},
	}

	// Load configuration from Azure App Configuration
	appConfig, err := azureappconfiguration.Load(ctx, authOptions, options)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	return appConfig, nil
}

func main() {
	ctx := context.Background()

	// Print welcome message
	fmt.Println("=== Azure App Configuration Feature Flags Console Demo ===")
	fmt.Println("This application demonstrates feature flag evaluation with dynamic refresh.")
	fmt.Println("Make sure to set the AZURE_APPCONFIG_CONNECTION_STRING environment variable.")
	fmt.Println("You can toggle the 'Beta' feature flag in the Azure portal to see real-time updates.")
	fmt.Println()

	// Load Azure App Configuration
	appConfig, err := loadAzureAppConfiguration(ctx)
	if err != nil {
		log.Fatalf("Error loading Azure App Configuration: %v", err)
	}

	// Create feature flag provider
	featureFlagProvider, err := azappconfig.NewFeatureFlagProvider(appConfig)
	if err != nil {
		log.Fatalf("Error creating feature flag provider: %v", err)
	}

	// Create feature manager
	featureManager, err := featuremanagement.NewFeatureManager(featureFlagProvider, nil)
	if err != nil {
		log.Fatalf("Error creating feature manager: %v", err)
	}

	// Monitor the Beta feature flag
	fmt.Println("Monitoring 'Beta' feature flag (press Ctrl+C to exit):")
	fmt.Println("Toggle the Beta feature flag in Azure portal to see real-time updates...")
	fmt.Println()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Refresh configuration to get latest feature flag settings
			if err := appConfig.Refresh(ctx); err != nil {
				log.Printf("Error refreshing configuration: %v", err)
				continue
			}

			// Evaluate the Beta feature flag
			isEnabled, err := featureManager.IsEnabled("Beta")
			if err != nil {
				log.Printf("Error checking if Beta feature is enabled: %v", err)
				continue
			}

			// Print timestamp and feature status
			timestamp := time.Now().Format("15:04:05")
			fmt.Printf("[%s] Beta is enabled: %t\n", timestamp, isEnabled)

		case <-ctx.Done():
			fmt.Println("\nShutting down...")
			return
		}
	}
}

