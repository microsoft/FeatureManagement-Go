package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "strings"

    "github.com/Azure/AppConfiguration-GoProvider/azureappconfiguration"
    "github.com/gin-contrib/sessions"
    "github.com/gin-contrib/sessions/cookie"
    "github.com/gin-gonic/gin"
    "github.com/microsoft/Featuremanagement-Go/featuremanagement"
    "github.com/microsoft/Featuremanagement-Go/featuremanagement/providers/azappconfig"
)

type Quote struct {
    Message string `json:"message"`
    Author  string `json:"author"`
}

type WebApp struct {
    featureManager *featuremanagement.FeatureManager
    appConfig      *azureappconfiguration.AzureAppConfiguration
    quotes         []Quote
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

    // Initialize quotes
    quotes := []Quote{
        {
            Message: "You cannot change what you are, only what you do.",
            Author:  "Philip Pullman",
        },
    }

    // Create web app
    app := &WebApp{
        featureManager: featureManager,
        appConfig:      appConfig,
        quotes:         quotes,
    }

    // Setup Gin with default middleware (Logger and Recovery)
    r := gin.Default()

    // Setup routes
    app.setupRoutes(r)

    // Start server
    if err := r.Run(":8080"); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }

    fmt.Println("Starting Quote of the Day server on http://localhost:8080")
    fmt.Println("Open http://localhost:8080 in your browser")
    fmt.Println()

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

    // Load HTML templates
    r.LoadHTMLGlob("templates/*.html")
    // Routes
    r.GET("/", app.homeHandler)
    r.GET("/login", app.loginPageHandler)
    r.POST("/login", app.loginHandler)
    r.GET("/logout", app.logoutHandler)
}

// Home page handler
func (app *WebApp) homeHandler(c *gin.Context) {
    session := sessions.Default(c)
    username := session.Get("username")
    quote := app.quotes[0]

    var greetingMessage string
    var targetingContext featuremanagement.TargetingContext
    if username != nil {
        // Create targeting context for the user
        targetingContext = createTargetingContext(username.(string))

        // Get the Greeting variant for the current user
        if variant, err := app.featureManager.GetVariant("Greeting", targetingContext); err != nil {
            log.Printf("Error getting Greeting variant: %v", err)
        } else if variant.ConfigurationValue != nil {
            // Extract the greeting message from the variant configuration
            if configValue, ok := variant.ConfigurationValue.(string); ok {
                greetingMessage = configValue
            }
        }
    }

    c.HTML(http.StatusOK, "index.html", gin.H{
        "title":           "Quote of the Day",
        "user":            username,
        "greetingMessage": greetingMessage,
        "quote":           quote,
    })
}

func (app *WebApp) loginPageHandler(c *gin.Context) {
    c.HTML(http.StatusOK, "login.html", gin.H{
        "title": "Login - Quote of the Day",
    })
}

func (app *WebApp) loginHandler(c *gin.Context) {
    email := strings.TrimSpace(c.PostForm("email"))

    // Basic validation
    if email == "" {
        c.HTML(http.StatusOK, "login.html", gin.H{
            "title": "Login - Quote of the Day",
            "error": "Email cannot be empty",
        })
        return
    }

    if !strings.Contains(email, "@") {
        c.HTML(http.StatusOK, "login.html", gin.H{
            "title": "Login - Quote of the Day",
            "error": "Please enter a valid email address",
        })
        return
    }

    // Store email in session
    session := sessions.Default(c)
    session.Set("username", email)
    if err := session.Save(); err != nil {
        log.Printf("Error saving session: %v", err)
    }

    c.Redirect(http.StatusFound, "/")
}

func (app *WebApp) logoutHandler(c *gin.Context) {
    session := sessions.Default(c)
    session.Clear()
    if err := session.Save(); err != nil {
        log.Printf("Error saving session: %v", err)
    }
    c.Redirect(http.StatusFound, "/")
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
            domain := parts[1]
            targetingContext.Groups = append(targetingContext.Groups, domain) // Add domain as group
        }
    }

    return targetingContext
}