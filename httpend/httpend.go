// Package httpend is responsible for the HTTP endpoint for groupplan
package httpend

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/discord"
	"github.com/wallnutkraken/groupplan/config"
	"github.com/wallnutkraken/groupplan/userman"
)

// Endpoint is the object used to start and handle the HTTP endpoint
type Endpoint struct {
	port    uint
	router  *gin.Engine
	userMan *userman.Manager

	authGroup *gin.RouterGroup
}

// New creates a new instance of the HTTP endpoint with the given port
func New(cfg config.AppSettings, userMan *userman.Manager) Endpoint {
	e := Endpoint{
		port:    cfg.Port,
		router:  gin.Default(),
		userMan: userMan,
	}
	// Start the goth discord provider
	goth.UseProviders(discord.New(cfg.DiscordKey, cfg.DiscordSecret, "http://fastvote.online/auth/discord/callback", discord.ScopeIdentify, discord.ScopeEmail))
	gothic.Store = sessions.NewFilesystemStore(os.TempDir(), []byte("goth-example"))

	// Add the HTML template Glob(?)
	e.router.LoadHTMLGlob("frontend/*/*/*")

	// Add the groups
	e.authGroup = e.router.Group("auth")

	// Auth group methods
	e.authGroup.GET(":provider", e.StartAuth)
	e.authGroup.GET(":provider/callback", e.AuthCallback)

	return e
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
		fmt.Println(err.Error())
		return
	}
	ctx.HTML(http.StatusOK, "discord.html", user)
}

// Start starts listening, this is a blocking call
func (e Endpoint) Start() error {
	return e.router.Run(fmt.Sprintf(":%d", e.port))
}
