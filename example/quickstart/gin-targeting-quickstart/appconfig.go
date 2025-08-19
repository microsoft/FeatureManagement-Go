package main

import (
    "context"
    "log"
    "os"

    "github.com/Azure/AppConfiguration-GoProvider/azureappconfiguration"
    "github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

func loadAzureAppConfiguration(ctx context.Context) (*azureappconfiguration.AzureAppConfiguration, error) {
    // Get the endpoint from environment variable
    endpoint := os.Getenv("AZURE_APPCONFIG_ENDPOINT")
    if endpoint == "" {
        log.Fatal("AZURE_APPCONFIG_ENDPOINT environment variable is not set")
    }

    // Create a credential using DefaultAzureCredential
    credential, err := azidentity.NewDefaultAzureCredential(nil)
    if err != nil {
        log.Fatalf("Failed to create credential: %v", err)
    }

    // Set up authentication options with endpoint and credential
    authOptions := azureappconfiguration.AuthenticationOptions{
        Endpoint:   endpoint,
        Credential: credential,
    }

    // Set up options to enable feature flags
    options := &azureappconfiguration.Options{
        FeatureFlagOptions: azureappconfiguration.FeatureFlagOptions{
            Enabled: true,
            RefreshOptions: azureappconfiguration.RefreshOptions{
                Enabled: true,
            },
        },
    }

    // Load configuration from Azure App Configuration
    appConfig, err := azureappconfiguration.Load(ctx, authOptions, options)
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    return appConfig, nil
}