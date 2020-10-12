// Package userman is the users manager, it's responsible for user operations with the data layer
package userman

import (
	"fmt"

	"github.com/wallnutkraken/groupplan/groupdata/users"
)

// Manager is responsible for user operations with the data layer
type Manager struct {
	users UserHandler
}

// UserHandler is the interface for what methods the user persistency layer should provide UserMan
type UserHandler interface {
	GetProviders() ([]users.AuthenticationProvider, error)
	GetProvider(name string) (users.AuthenticationProvider, error)
	GetOrCreateUser(email, avatarURL, displayName string) (users.User, error)
	UserAuthorizedWith(user users.User, provider users.AuthenticationProvider, identifier string) (users.UserAuthPoint, error)
}

// New creates a new instance of the user manager
func New(userData UserHandler) *Manager {
	return &Manager{
		users: userData,
	}
}

// Authenticate takes a user that was authenticated and saves
// them to the database if they're new
//
// Errors from this function MUST be handled internally (without sending them to the consumer)
func (m *Manager) Authenticate(email, avatarURL, provider, identifier, displayName string) (users.User, error) {
	prov, err := m.users.GetProvider(provider)
	if err != nil {
		return users.User{}, err
	}
	// Save them to the database
	user, err := m.users.GetOrCreateUser(email, avatarURL, displayName)
	if err != nil {
		return user, fmt.Errorf("failed creating/getting user from db: %w", err)
	}
	// Add the authorization point
	authPoint, err := m.users.UserAuthorizedWith(user, prov, identifier)
	if err != nil {
		return user, fmt.Errorf("failed creating authorization point: %w", err)
	}
	// Add authpoint to user object, we just saved it to db with no error. This should be fine.
	// Famous last words.
	user.AuthPoints = []users.UserAuthPoint{authPoint}

	// And return the user!
	return user, nil
}

// GetAuthenticatedUser returns an existing user based on their email address
func (m *Manager) GetAuthenticatedUser(email string) (users.User, error) {
	return m.users.GetOrCreateUser(email, "", "")
}
