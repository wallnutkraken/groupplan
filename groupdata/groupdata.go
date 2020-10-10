// Package groupdata is responsible for the data access for groupplan
package groupdata

import (
	"fmt"

	"github.com/wallnutkraken/groupplan/groupdata/users"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var dbPath string

func init() {
	// Get the database path on i
}

// Data is the wrapper struct for our GORM object, and contains methods to interact with the persistent storage.
type Data struct {
	db *gorm.DB
}

// New creates a new instance of the Data access object
func New(confPath string) (Data, error) {
	// Open gorm with the given data path
	db, err := gorm.Open(sqlite.Open(confPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return Data{}, fmt.Errorf("Failed opening sqlite file at [%s]: %w", confPath, err)
	}
	// Wrap the gorm object in our Data object
	data := Data{
		db: db,
	}
	// And call migrate to automigrate
	err = data.migrate()
	if err != nil {
		return data, fmt.Errorf("Failed migrating database: %w", err)
	}
	return data, nil
}

// Users returns the users handler
func (d Data) Users() users.UserHandler {
	return users.New(d.db)
}

// migrate collects all the database data types and calls gorm's AutoMigrate method
// to migrate the schema to the database
func (d Data) migrate() error {
	allDataTypes := []interface{}{}
	allDataTypes = append(allDataTypes, users.AllTypes()...)

	return d.db.AutoMigrate(allDataTypes...)
}
