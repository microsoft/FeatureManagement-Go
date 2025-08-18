package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "strings"

    "github.com/gin-contrib/sessions"
    "github.com/gin-contrib/sessions/cookie"
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
            if err := app.appConfig.Refresh(context.Background()); err != nil {
                log.Printf("Error refreshing configuration: %v", err)
            }
        }()
        c.Next()
    }
}

func (app *WebApp) featureMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get current user from session
        session := sessions.Default(c)
        username := session.Get("username")

        var betaEnabled bool
        var targetingContext featuremanagement.TargetingContext
        if username != nil {
            // Evaluate Beta feature with targeting context
            var err error
            targetingContext = createTargetingContext(username.(string))
            betaEnabled, err = app.featureManager.IsEnabledWithAppContext("Beta", targetingContext)
            if err != nil {
                log.Printf("Error checking Beta feature with targeting: %v", err)
            }
        }

        c.Set("betaEnabled", betaEnabled)
        c.Set("user", username)
        c.Set("targetingContext", targetingContext)
        c.Next()
    }
}

// Helper function to create TargetingContext
func createTargetingContext(userID string) featuremanagement.TargetingContext {
    targetingContext := featuremanagement.TargetingContext{
        UserID: userID,
        Groups: []string{},
    }

    if strings.Contains(userID, "@") {
        parts := strings.Split(userID, "@")
        if len(parts) == 2 {
            targetingContext.Groups = append(targetingContext.Groups, parts[1]) // Add domain as group
        }
    }

    return targetingContext
}

func (app *WebApp) setupRoutes(r *gin.Engine) {
    // Setup sessions
    store := cookie.NewStore([]byte("secret-key-change-in-production"))
    store.Options(sessions.Options{
        MaxAge:   3600, // 1 hour
        HttpOnly: true,
        Secure:   false, // Set to true in production with HTTPS
    })
    r.Use(sessions.Sessions("session", store))


    r.Use(app.refreshMiddleware())
    r.Use(app.featureMiddleware())

    // Load HTML templates
    r.LoadHTMLGlob("templates/*.html")

    // Routes
    r.GET("/", app.homeHandler)
    r.GET("/beta", app.betaHandler)
    r.GET("/login", app.loginPageHandler)
    r.POST("/login", app.loginHandler)
    r.GET("/logout", app.logoutHandler)
}

// Home page handler
func (app *WebApp) homeHandler(c *gin.Context) {
    betaEnabled := c.GetBool("betaEnabled")
    user := c.GetString("user")

    c.HTML(http.StatusOK, "index.html", gin.H{
        "title":       "TestFeatureFlags",
        "betaEnabled": betaEnabled,
        "user":        user,
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

func (app *WebApp) loginPageHandler(c *gin.Context) {
    c.HTML(http.StatusOK, "login.html", gin.H{
        "title": "Login",
    })
}

func (app *WebApp) loginHandler(c *gin.Context) {
    username := c.PostForm("username")

    // Basic validation - ensure username is not empty
    if strings.TrimSpace(username) == "" {
        c.HTML(http.StatusOK, "login.html", gin.H{
            "title": "Login",
            "error": "Username cannot be empty",
        })
        return
    }

    // Store username in session - any valid username is accepted
    session := sessions.Default(c)
    session.Set("username", username)
    session.Save()
    c.Redirect(http.StatusFound, "/")
}

func (app *WebApp) logoutHandler(c *gin.Context) {
    session := sessions.Default(c)
    session.Clear()
    session.Save()
    c.Redirect(http.StatusFound, "/")
}

func main() {
    // Load Azure App Configuration
    appConfig, err := loadAzureAppConfiguration(context.Background())
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

    // Setup Gin with default middleware (Logger and Recovery)
    r := gin.Default()

    // Setup routes
    app.setupRoutes(r)

    // Start server
    if err := r.Run(":8080"); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }

    fmt.Println("Starting server on http://localhost:8080")
    fmt.Println("Open http://localhost:8080 in your browser")
    fmt.Println()
}