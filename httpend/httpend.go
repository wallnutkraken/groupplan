// Package httpend is responsible for the HTTP endpoint for groupplan
package httpend

import (
	"fmt"
	"net/http"

	"github.com/wallnutkraken/groupplan/groupdata"
	"github.com/wallnutkraken/groupplan/httpend/plan"
	"github.com/wallnutkraken/groupplan/planman"

	"golang.org/x/crypto/acme/autocert"

	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"github.com/wallnutkraken/groupplan/config"
	"github.com/wallnutkraken/groupplan/httpend/userauth"
	"github.com/wallnutkraken/groupplan/userman"
)

// Endpoint is the object used to start and handle the HTTP endpoint
type Endpoint struct {
	hostname    string
	router      *gin.Engine
	authHandler *userauth.Handler
	planHanlder *plan.Handler
}

// New creates a new instance of the HTTP endpoint with the given port
func New(cfg config.AppSettings, db groupdata.Data) Endpoint {
	e := Endpoint{
		router:   gin.Default(),
		hostname: cfg.Hostname,
	}
	// Initialize the sub-handlers
	e.authHandler = userauth.New(e.router, userman.New(db.Users()), cfg)
	e.planHanlder = plan.New(e.router, e.authHandler, planman.New(db.Plans()))

	e.router.StaticFS("static", http.Dir("frontend/static"))
	// And HTML endpoint methods
	e.router.GET("/", e.Index)

	// Ping handler
	e.router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	return e
}

// Index returns the HTML index document
func (e Endpoint) Index(ctx *gin.Context) {
	_, err := e.authHandler.GetJWT(ctx)
	if err != nil {
		// User not authed, send them to login
		fmt.Println("returning login")
		ctx.File("frontend/login.html")
		return
	}
	// User authed, give them the dashboard
	fmt.Println("Returning dashboard")
	ctx.File("frontend/dashboard.html")
}

// Start starts listening, this is a blocking call
func (e Endpoint) Start() error {
	return autotls.Run(e.router, "www.fastvote.online", e.hostname)
	return autotls.RunWithManager(e.router, &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(e.hostname, fmt.Sprintf("www.%s", e.hostname)),
	})
}
