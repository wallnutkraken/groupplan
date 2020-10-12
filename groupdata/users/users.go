// Package users is responsible for user data
package users

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var supportedProviders = []string{
	"discord",
}

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

// GetProviders returns all the supported authentication providers from the database
func (u UserHandler) GetProviders() ([]AuthenticationProvider, error) {
	prov := []AuthenticationProvider{}
	if err := u.db.Find(&prov).Error; err != nil {
		return nil, fmt.Errorf("could not get a list of authentication providers: %w", err)
	}
	return prov, nil
}

// GetProvider returns a single provider with a matching name
func (u UserHandler) GetProvider(name string) (prov AuthenticationProvider, err error) {
	if err = u.db.Where(AuthenticationProvider{Name: name}).First(&prov).Error; err != nil {
		err = fmt.Errorf("failed finding provider: %w", err)
	}
	return
}

// GetOrCreateUser tries to get the record for the given user Email address. If that fails, it
// will create a new user with that email and return it. It also takes an avatarURL parameter.
// As any time we would be authenticating a user, we'd have their avatar URL, we'll past it here
// for the purposes of creating a user. If the user already exists, this URL will not change the one
// stored in the database.
func (u UserHandler) GetOrCreateUser(email, avatarURL, displayName string) (User, error) {
	usr := User{
		Email:             email,
		ProfilePictureURL: avatarURL,
		DisplayName:       displayName,
	}
	if err := u.db.Preload(clause.Associations).Where(User{Email: email}).FirstOrCreate(&usr).Error; err != nil {
		return usr, fmt.Errorf("failed getting/creating user with email [%s]: %w", email, err)
	}

	return usr, nil
}

// UserAuthorizedWith checks if the user is authorized with a given authentication provider.
// If not, it will create an authroization entry with the data given
func (u UserHandler) UserAuthorizedWith(user User, provider AuthenticationProvider, identifier string) (UserAuthPoint, error) {
	authPt := UserAuthPoint{
		UserID:     user.ID,
		ProviderID: provider.ID,
		Identifier: identifier,
	}
	if err := u.db.Preload(clause.Associations).Where(UserAuthPoint{UserID: user.ID, ProviderID: provider.ID}).FirstOrCreate(&authPt).Error; err != nil {
		return authPt, fmt.Errorf("failed getting/creating user auth point for user email [%s] provider [%s] and identifier [%s]: %w", user.Email, provider.Name, identifier, err)
	}

	return authPt, nil
}

// AuthenticationProvider contains information about an oauth provider
type AuthenticationProvider struct {
	ID   uint   `gorm:"primarykey"`
	Name string `gorm:"index"`
}

// User represents a single user
type User struct {
	gorm.Model
	Email             string `gorm:"unique"`
	DisplayName       string
	ProfilePictureURL string
	AuthPoints        []UserAuthPoint `gorm:"foreignKey:UserID"`
}

// UserAuthPoint contains information about a single point of authentication for a user
type UserAuthPoint struct {
	ID         uint `gorm:"primarykey"`
	UserID     uint
	Identifier string
	Provider   AuthenticationProvider `gorm:"foreignkey:ProviderID"`
	ProviderID uint
}

// AllTypes returns all the gorm data types defined in this package, to be used with gorm.AutoMigrate
func AllTypes() []interface{} {
	return []interface{}{AuthenticationProvider{}, User{}, UserAuthPoint{}}
}

// Migrate ensures the necessary minimum data exists in the database
func Migrate(db *gorm.DB) error {
	authProviders := []AuthenticationProvider{}
	if err := db.Find(&authProviders).Error; err != nil {
		return fmt.Errorf("failed getting auth providers: %w", err)
	}
	// Create an array of missing providers so we can add them later
	missing := []AuthenticationProvider{}
	for _, provider := range supportedProviders {
		found := false
		for _, existing := range authProviders {
			if existing.Name == provider {
				found = true
			}
		}
		if !found {
			missing = append(missing, AuthenticationProvider{Name: provider})
		}
	}

	// Save each missing provider
	for _, prov := range missing {
		if err := db.Create(&prov).Error; err != nil {
			return fmt.Errorf("failed saving provider [%s] to database: %w", prov.Name, err)
		}
	}

	// All done!
	return nil
}
