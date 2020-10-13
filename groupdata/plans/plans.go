// Package plans is responsible for plan storage in the persistence layer
package plans

import (
	"fmt"
	"time"

	"gorm.io/gorm/clause"

	"github.com/wallnutkraken/groupplan/groupdata/users"
	"gorm.io/gorm"
)

// PlanHandler is the data sub-hanlder for the plans package, dealing with data relevant to plans
type PlanHandler struct {
	db *gorm.DB
}

// New creates a new instance of the PlanHandler with the given gorm DB instance
func New(db *gorm.DB) PlanHandler {
	return PlanHandler{
		db: db,
	}
}

// AllTypes returns all the gorm data types defined in this package, to be used with gorm.AutoMigrate
func AllTypes() []interface{} {
	return []interface{}{PlanEntry{}, Plan{}}
}

// CreatePlan creates a new entry in the database for the given plan.
// It does not create entries for the `Entries` element or its children.
func (p *PlanHandler) CreatePlan(plan *Plan) error {
	if err := p.db.Create(plan).Error; err != nil {
		return fmt.Errorf("failed creating plan: %w", err)
	}
	return nil
}

// GetPlan returns an existing Plan by the identifier
func (p *PlanHandler) GetPlan(identifier string) (plan Plan, err error) {
	if err = p.db.Preload(clause.Associations).Where(Plan{Identifier: identifier}).First(&plan).Error; err != nil {
		err = fmt.Errorf("failed getting plan [%s]: %w", identifier, err)
	}
	return
}

// DeletePlan deletes the provided plan from the database
func (p *PlanHandler) DeletePlan(plan Plan) error {
	if err := p.db.Delete(&plan).Error; err != nil {
		return fmt.Errorf("failed deleting plan: %w", err)
	}
	return nil
}

// AddEntry creates a new plan availability entry for a user with a given time range,
// then adds the created object to the provided Plan pointer.
func (p *PlanHandler) AddEntry(plan *Plan, user users.User, availFrom, availTo int64) (PlanEntry, error) {
	// Create the entry object
	entry := PlanEntry{
		User:          user,
		UserID:        user.ID,
		PlanID:        plan.ID,
		StartTimeUnix: availFrom,
		EndTimeUnix:   availTo,
	}
	// Write it to the database
	if err := p.db.Create(&entry).Error; err != nil {
		return entry, fmt.Errorf("failed creating plan entry: %w", err)
	}

	// Created successfully. Add it to plan and return it.
	plan.Entries = append(plan.Entries, entry)

	return entry, nil
}

// Plan represents a plan in the data layer
type Plan struct {
	gorm.Model
	Owner      users.User  `gorm:"foreignkey:OwnerID"`
	OwnerID    uint        `gorm:"not null"`
	Identifier string      `gorm:"index;not null"`
	Title      string      `gorm:"not null"`
	FromDate   time.Time   `gorm:"not null"`
	ToDate     time.Time   `gorm:"not null"`
	Entries    []PlanEntry `gorm:"foreignkey:PlanID"`
}

// PlanEntry represents one user's entries in a single plan
type PlanEntry struct {
	gorm.Model
	User          users.User `gorm:"foreignkey:UserID"`
	UserID        uint       `gorm:"not null"`
	PlanID        uint       `gorm:"not null"`
	StartTimeUnix int64      `gorm:"not null"`
	EndTimeUnix   int64      `gorm:"not null"`
}
