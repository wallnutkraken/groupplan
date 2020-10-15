// Package httpend is responsible for the HTTP endpoint for groupplan
package httpend

import (
	"fmt"
	"io/ioutil"
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

	loginHTML     []byte
	dashboardHTML []byte
}

// loadHTML loads the login and dashboard HTML files into memory
func (e *Endpoint) loadHTML() error {
	login, err := ioutil.ReadFile("frontend/login.html")
	if err != nil {
		return fmt.Errorf("failed reading the login file: %w", err)
	}
	dashboard, err := ioutil.ReadFile("frontend/dashboard.html")
	if err != nil {
		return fmt.Errorf("failed reading the dashboard file: %w", err)
	}

	// Put the file info into the endpoint struct and return nil
	e.loginHTML = login
	e.dashboardHTML = dashboard
	return nil
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

	// Load the dashboard and login HTML files, as we'll be serving them from memory
	e.loadHTML()

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
		ctx.Data(http.StatusOK, "text/html", e.loginHTML)
		return
	}
	// User authed, give them the dashboard
	fmt.Println("Returning dashboard")
	ctx.Data(http.StatusOK, "text/html", e.dashboardHTML)
}

// Start starts listening, this is a blocking call
func (e Endpoint) Start() error {
	return autotls.Run(e.router, "www.fastvote.online", e.hostname)
	return autotls.RunWithManager(e.router, &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(e.hostname, fmt.Sprintf("www.%s", e.hostname)),
	})
}
