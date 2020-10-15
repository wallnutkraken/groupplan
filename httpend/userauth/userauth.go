// Package userauth contains the User/Authentication HTTP endpoint logic/router group
package userauth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/discord"
	"github.com/sirupsen/logrus"
	"github.com/wallnutkraken/groupplan/config"
	"github.com/wallnutkraken/groupplan/groupdata/users"
	"github.com/wallnutkraken/groupplan/userman"
)

const (
	authCookie = "groupplan_jwt"
)

// Handler is the object responsible for the /auth endpoint
type Handler struct {
	hostname  string
	group     *gin.RouterGroup
	userMan   *userman.Manager
	jwtSecret string //todo: memguard

	expireAfterSeconds int64
}

// Authenticator is the interface for objects that allow JWT authentication for other (non-sign in) endpoints
type Authenticator interface {
	GetJWT(ctx *gin.Context) (users.User, error)
}

// GroupPlanClaims is the JWT authentication claims object for GroupPlan
type GroupPlanClaims struct {
	jwt.StandardClaims
	Email       string `json:"email"`
	AvatarURL   string `json:"pfp"`
	DisplayName string `json:"name"`
}

// generateJWTSecret generates a secure random string for use with JWT
func generateJWTSecret() string {
	// Generate 256 bytes of randomness
	randBytes := make([]byte, 256)
	written, err := rand.Read(randBytes)
	if err != nil {
		logrus.WithError(err).Fatal("Random number generator failed creating a secure JWT string")
	}
	// Encode to base64 and return the string
	return base64.StdEncoding.EncodeToString(randBytes[:written])
}

// New creates a new instance of the authentication Handler
func New(router *gin.Engine, userH *userman.Manager, cfg config.AppSettings) *Handler {
	handler := &Handler{
		group:              router.Group("auth"),
		userMan:            userH,
		expireAfterSeconds: 3600 * 24,
		hostname:           cfg.Hostname,
		jwtSecret:          generateJWTSecret(),
	}
	// Start the goth discord provider
	goth.UseProviders(discord.New(cfg.DiscordKey, cfg.DiscordSecret, fmt.Sprintf("https://%s/auth/discord/callback", cfg.Hostname), discord.ScopeIdentify, discord.ScopeEmail))

	// Auth group methods
	handler.group.GET(":provider", handler.StartAuth)
	handler.group.GET(":provider/callback", handler.AuthCallback)

	return handler
}

// StartAuth is the endpoint to begin the authentication process
func (h Handler) StartAuth(ctx *gin.Context) {
	// Add the discord provider to the context so that gothic knows what we're trying to authenticate with
	req := gothic.GetContextWithProvider(ctx.Request, ctx.Param("provider"))

	// First, try to get the user without re-authenticating
	if user, err := gothic.CompleteUserAuth(ctx.Writer, req); err == nil {
		fmt.Printf("%+v\n", user)
	} else {
		gothic.BeginAuthHandler(ctx.Writer, req)
	}
}

// GetJWT takes the request context and returns a parsed claim object if the authentication is valid
func (h Handler) GetJWT(ctx *gin.Context) (users.User, error) {
	jwtRaw, err := ctx.Cookie(authCookie)
	if err != nil {
		// ErrNoCookie
		return users.User{}, fmt.Errorf("no authentication cookie: %w", err)
	}
	cl := GroupPlanClaims{}
	parsed, err := jwt.ParseWithClaims(jwtRaw, &cl, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method [%s]", token.Header["alg"])
		}
		return []byte(h.jwtSecret), nil
	})
	if err != nil {
		return users.User{}, fmt.Errorf("failed parsing jwt: %w", err)
	}
	if !parsed.Valid {
		return users.User{}, errors.New("invalid jwt")
	}

	// Get the user data from the database about this user
	return h.userMan.GetAuthenticatedUser(cl.Email)
}

// AuthCallback is the HTTP endpoint for the Discord authorization callback
func (h Handler) AuthCallback(ctx *gin.Context) {
	provider := ctx.Param("provider")
	req := gothic.GetContextWithProvider(ctx.Request, provider)

	user, err := gothic.CompleteUserAuth(ctx.Writer, req)
	if err != nil {
		logrus.WithError(err).Error("Failed completing user authentication")
		ctx.AbortWithStatus(http.StatusInternalServerError) // todo error here
		return
	}

	// Create a JWT for the user
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, GroupPlanClaims{
		Email:       user.Email,
		AvatarURL:   user.AvatarURL,
		DisplayName: user.Name,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + h.expireAfterSeconds,
		},
	})
	// Sign the token
	signed, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		logrus.WithError(err).Error("Failed signing JWT")
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Create user info. We don't actually care about the returned object here, because err being nil
	// guarantees we actually created the user
	_, err = h.userMan.Authenticate(user.Email, user.AvatarURL, provider, user.UserID, user.Name)
	if err != nil {
		// Aight, err wasn't nil
		logrus.WithError(err).Errorf("Failed saving/getting user after authentication with email [%s] and provider [%s]", user.Email, provider)
		ctx.AbortWithStatus(http.StatusInternalServerError) // Todo: errorpage
		return
	}

	ctx.SetCookie(authCookie, signed, int(h.expireAfterSeconds), "", h.hostname, true, false)

	ctx.Redirect(http.StatusFound, "https://fastvote.online")
}
