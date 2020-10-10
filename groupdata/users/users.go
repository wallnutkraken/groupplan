// Package users is responsible for user data
package users

import "gorm.io/gorm"

// UserHandler is the data sub-handler for the users package, dealing with data relevant to users
type UserHandler struct {
	db *gorm.DB
}

// New creates a new instance of the userHandler with the given gorm DB pointer
func New(db *gorm.DB) UserHandler {
	return UserHandler{
		db: db,
	}
}

// User represents a single user
type User struct {
	gorm.Model
	DiscordID string
}

// AllTypes returns all the gorm data types defined in this package, to be used with gorm.AutoMigrate
func AllTypes() []interface{} {
	return []interface{}{User{}}
}
