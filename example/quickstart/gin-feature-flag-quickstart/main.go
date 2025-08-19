package main

import (
    "context"
    "fmt"
    "log"
    "net/http"

    "github.com/Azure/AppConfiguration-GoProvider/azureappconfiguration"
    "github.com/gin-gonic/gin"
    "github.com/microsoft/Featuremanagement-Go/featuremanagement"
    "github.com/microsoft/Featuremanagement-Go/featuremanagement/providers/azappconfig"
)

type WebApp struct {
    featureManager *featuremanagement.FeatureManager
    appConfig      *azureappconfiguration.AzureAppConfiguration
}

func (app *WebApp) refreshMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        go func() {
            ctx := context.Background()
            if err := app.appConfig.Refresh(ctx); err != nil {
                log.Printf("Error refreshing configuration: %v", err)
            }
        }()
        c.Next()
    }
}

func (app *WebApp) featureMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Check if Beta feature is enabled
        betaEnabled, err := app.featureManager.IsEnabled("Beta")
        if err != nil {
            log.Printf("Error checking Beta feature: %v", err)
            betaEnabled = false
        }

        // Store feature flag status for use in templates
        c.Set("betaEnabled", betaEnabled)
        c.Next()
    }
}

func (app *WebApp) setupRoutes(r *gin.Engine) {
    r.Use(app.refreshMiddleware())
    r.Use(app.featureMiddleware())

    // Load HTML templates
    r.LoadHTMLGlob("templates/*.html")

    // Routes
    r.GET("/", app.homeHandler)
    r.GET("/beta", app.betaHandler)
}

// Home page handler
func (app *WebApp) homeHandler(c *gin.Context) {
    betaEnabled := c.GetBool("betaEnabled")

    c.HTML(http.StatusOK, "index.html", gin.H{
        "title":       "Feature Management Example App",
        "betaEnabled": betaEnabled,
    })
}

// Beta page handler
func (app *WebApp) betaHandler(c *gin.Context) {
    betaEnabled := c.GetBool("betaEnabled")

    if !betaEnabled {
        return
    }

    c.HTML(http.StatusOK, "beta.html", gin.H{
        "title": "Beta Page",
    })
}

func main() {
    ctx := context.Background()

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

    // Create web app
    app := &WebApp{
        featureManager: featureManager,
        appConfig:      appConfig,
    }

    // Set up Gin with default middleware (Logger and Recovery)
    r := gin.Default()

    // Set up routes
    app.setupRoutes(r)

    // Start server
    fmt.Println("Starting server on http://localhost:8080")
    fmt.Println("Open http://localhost:8080 in your browser")
    fmt.Println("Toggle the 'Beta' feature flag in Azure portal to see changes")
    fmt.Println()

    if err := r.Run(":8080"); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}