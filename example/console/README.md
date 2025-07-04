# Feature Management Console App Example

This is a simple Go console application that demonstrates how to use Azure App Configuration feature flags with the Go Feature Management library.

## Overview

The application:

- Loads feature flags from Azure App Configuration
- Feature flag evaluation using the Go Feature Management library
- Automatically refreshes the feature flags when changed in Azure App Configuration

## Prerequisites

- An Azure account with an active subscription
- An Azure App Configuration store
- A feature flag named "Beta" in your App Configuration store
- Go 1.23 or later

## Running the Example

1. **Create a feature flag in Azure App Configuration:**

   Add a feature flag called *Beta* to the App Configuration store and leave **Label** and **Description** with their default values. For more information about how to add feature flags to a store using the Azure portal or the CLI, go to [Create a feature flag](https://learn.microsoft.com/azure/azure-app-configuration/manage-feature-flags?tabs=azure-portal#create-a-feature-flag).
   
2. **Set environment variable:**

   **Windows PowerShell:**
   ```powershell
   $env:AZURE_APPCONFIG_CONNECTION_STRING = "your-connection-string"
   ```

   **Windows Command Prompt:**
   ```cmd
   setx AZURE_APPCONFIG_CONNECTION_STRING "your-connection-string"
   ```

   **Linux/macOS:**
   ```bash
   export AZURE_APPCONFIG_CONNECTION_STRING="your-connection-string"
   ```

## Running the Application

```bash
go run main.go
```

## Testing the Feature Flag

1. Start the application - you should see `Beta is enabled: false`
2. Go to Azure portal → App Configuration → Feature manager
3. Enable the "Beta" feature flag
4. Wait a few seconds and observe the console output change to `Beta is enabled: true`
5. Disable the flag again to see it change back to `false`
