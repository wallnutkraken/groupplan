// Package httpend is responsible for the HTTP endpoint for groupplan
package httpend

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/discord"
	"github.com/sirupsen/logrus"
	"github.com/wallnutkraken/groupplan/config"
	"github.com/wallnutkraken/groupplan/userman"
)

// Endpoint is the object used to start and handle the HTTP endpoint
type Endpoint struct {
	router  *gin.Engine
	userMan *userman.Manager

	authGroup *gin.RouterGroup
}

const (
	authCookie = "groupplan_authorization_details"
)

// New creates a new instance of the HTTP endpoint with the given port
func New(cfg config.AppSettings, userMan *userman.Manager) Endpoint {
	e := Endpoint{
		router:  gin.Default(),
		userMan: userMan,
	}
	// Start the goth discord provider
	goth.UseProviders(discord.New(cfg.DiscordKey, cfg.DiscordSecret, fmt.Sprintf("https://%s/auth/discord/callback", cfg.Hostname), discord.ScopeIdentify, discord.ScopeEmail))

	// Add the HTML template Glob(?)
	e.router.LoadHTMLGlob("frontend/*/*.html")
	// And HTML endpoint methods
	e.router.GET("", e.Index)

	// Ping handler
	e.router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Add the groups
	e.authGroup = e.router.Group("auth")

	// Auth group methods
	e.authGroup.GET(":provider", e.StartAuth)
	e.authGroup.GET(":provider/callback", e.AuthCallback)

	return e
}

// Index returns the HTML index document
func (e Endpoint) Index(ctx *gin.Context) {
	// First, try and get the user info. If that fails, give them login.html
	// Otherwise, give them the dashboard.html
	var user goth.User

	// Get the auth cookie
	cookie, err := ctx.Cookie(authCookie)
	if err != nil {
		// Just send them to the login
		ctx.HTML(http.StatusUnauthorized, "login.html", nil)
		return
	}
	// Decode the cookie
	if err := json.Unmarshal([]byte(cookie), &user); err != nil {
		// Failed to do the cookie, just go to login again
		ctx.HTML(http.StatusUnauthorized, "login.html", nil)
		return
	}
	ctx.HTML(http.StatusOK, "dashboard.html", user)
}

// StartAuth is the endpoint to begin the authentication process
func (e Endpoint) StartAuth(ctx *gin.Context) {
	// Add the discord provider to the context so that gothic knows what we're trying to authenticate with
	req := gothic.GetContextWithProvider(ctx.Request, ctx.Param("provider"))

	// First, try to get the user without re-authenticating
	if user, err := gothic.CompleteUserAuth(ctx.Writer, req); err == nil {
		fmt.Printf("%+v\n", user)
	} else {
		gothic.BeginAuthHandler(ctx.Writer, req)
	}
}

// AuthCallback is the HTTP endpoint for the Discord authorization callback
func (e Endpoint) AuthCallback(ctx *gin.Context) {
	req := gothic.GetContextWithProvider(ctx.Request, ctx.Param("provider"))

	user, err := gothic.CompleteUserAuth(ctx.Writer, req)
	if err != nil {
		logrus.WithError(err).Error("Failed completing user authentication")
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	// Add the Authorization cookie, first, jsonify the user, then put it in the cookie
	jsonData, err := json.Marshal(user)
	if err != nil {
		logrus.WithError(err).Errorf("Failed marshalling User to JSON: %+v", user)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	ctx.SetCookie(authCookie, string(jsonData), 3600*24, "", "fastvote.online", false, false)
	ctx.Redirect(http.StatusFound, "https://fastvote.online")
}

// Start starts listening, this is a blocking call
func (e Endpoint) Start() error {
	return autotls.Run(e.router, "fastvote.online")
}
