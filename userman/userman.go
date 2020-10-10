// Package userman is the users manager, it's responsible for user operations with the data layer
package userman

import "github.com/wallnutkraken/groupplan/groupdata/users"

// Manager is responsible for user operations with the data layer
type Manager struct {
	users users.UserHandler
}

// New creates a new instance of the user manager
func New(userData users.UserHandler) *Manager {
	return &Manager{
		users: userData,
	}
}
