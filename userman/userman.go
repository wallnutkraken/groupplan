// Package userman is the users manager, it's responsible for user operations with the data layer
package userman

import "github.com/wallnutkraken/groupplan/groupdata/users"

// Manager is responsible for user operations with the data layer
type Manager struct {
	users UserHandler
}

// UserHandler is the interface for what methods the user persistency layer should provide UserMan
type UserHandler interface {
	GetProviders() ([]users.AuthenticationProvider, error)
	GetOrCreateUser(email, avatarURL string) (users.User, error)
	UserAuthorizedWith(user users.User, provider users.AuthenticationProvider, identifier string) (users.UserAuthPoint, error)
}

// New creates a new instance of the user manager
func New(userData UserHandler) *Manager {
	return &Manager{
		users: userData,
	}
}

// Authenticate takes a
func (m *Manager) Authenticate() {

}
