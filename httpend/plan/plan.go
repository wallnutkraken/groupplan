// Package plan is responsible for all endpoints on the /plans resource
package plan

import (
	"github.com/gin-gonic/gin"
	"github.com/wallnutkraken/groupplan/httpend/userauth"
)

// Handler is the object responsible for the /plans endpoint
type Handler struct {
	group  *gin.RouterGroup
	auther userauth.Authenticator
}

// New creates a new instance of the plans handler
func New(router *gin.Engine, auth userauth.Authenticator) *Handler {
	handl := &Handler{
		group:  router.Group("plans"),
		auther: auth,
	}

	return handl
}
