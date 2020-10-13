// Package plans is responsible for plan storage in the persistence layer
package plans

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm/clause"

	"github.com/wallnutkraken/groupplan/groupdata/dataerror"
	"github.com/wallnutkraken/groupplan/groupdata/users"
	"gorm.io/gorm"
)

// PlanHandler is the data sub-hanlder for the plans package, dealing with data relevant to plans
type PlanHandler struct {
	db *gorm.DB
}

// New creates a new instance of the PlanHandler with the given gorm DB instance
func New(db *gorm.DB) *PlanHandler {
	return &PlanHandler{
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
	if err := plan.Validate(); err != nil {
		return fmt.Errorf("plan failed validation: %w", err)
	}
	if err := p.db.Create(plan).Error; err != nil {
		return fmt.Errorf("failed creating plan: %w", err)
	}
	return nil
}

// GetPlan returns an existing Plan by the identifier
func (p *PlanHandler) GetPlan(identifier string) (plan Plan, err error) {
	if err = p.db.Preload(clause.Associations).Where(Plan{Identifier: identifier}).First(&plan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = dataerror.ErrBasic("no such plan exists")
		}
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

// GetPlansByUser returns all plans by a given user
func (p *PlanHandler) GetPlansByUser(user users.User) ([]Plan, error) {
	plans := []Plan{}
	if err := p.db.Preload(clause.Associations).Where(Plan{OwnerID: user.ID}).Find(&plans).Error; err != nil {
		return nil, fmt.Errorf("failed getting plans for user ID [%d]: %w", user.ID, err)
	}

	return plans, nil
}

// AddEntry creates a new plan availability entry for a user with a given time range,
// then adds the created object to the provided Plan pointer.
func (p *PlanHandler) AddEntry(plan *Plan, user users.User, availFrom, durationSecs int64) (PlanEntry, error) {
	// Create the entry object
	entry := PlanEntry{
		User:            user,
		UserID:          user.ID,
		PlanID:          plan.ID,
		StartTimeUnix:   availFrom,
		DurationSeconds: durationSecs,
	}
	// Validate the entry
	if err := entry.Validate(); err != nil {
		return PlanEntry{}, fmt.Errorf("failed validating plan entry: %w", err)
	}
	// Check if it's within the bounds of its parent plan
	if plan.FromDateZeroHour().After(time.Unix(availFrom, 0)) {
		return PlanEntry{}, dataerror.ErrBasic("start at time cannot be before the plan start date")
	}
	if plan.EndDate().Before(time.Unix(availFrom, 0)) {
		return PlanEntry{}, dataerror.ErrBasic("start time can't be after plan end date")
	}
	if plan.EndDate().Before(time.Unix(availFrom+durationSecs, 0)) {
		return PlanEntry{}, dataerror.ErrBasic("this entry would end after the plan ends")
	}

	// Check if it overlaps with any current availability
	conflicts := []PlanEntry{}
	err := p.db.Where("((? >= start_time_unix AND ? <= start_time_unix+duration_seconds) OR (start_time_unix >= ? AND start_time_unix <= ?)) AND user_id = ? AND plan_id = ?",
		availFrom, availFrom, availFrom, availFrom+durationSecs, user.ID, plan.ID).Find(&conflicts).Error
	if err != nil {
		return PlanEntry{}, fmt.Errorf("failed checking for conflicting entries: %w", err)
	}
	if len(conflicts) != 0 {
		return PlanEntry{}, dataerror.ErrBasic("availability conflicts with another entry owned by the same user")
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
	Owner        users.User  `gorm:"foreignkey:OwnerID"`
	OwnerID      uint        `gorm:"not null"`
	Identifier   string      `gorm:"index;not null"`
	Title        string      `gorm:"not null"`
	FromDate     time.Time   `gorm:"not null"`
	DurationDays uint        `gorm:"not null"`
	Entries      []PlanEntry `gorm:"foreignkey:PlanID"`
}

// FromDateZeroHour takes the given start date and returns a time
// 0 seconds after the start of that date
func (p Plan) FromDateZeroHour() time.Time {
	y, m, d := p.FromDate.Date()
	zeroed, err := time.Parse("2006-1-2", fmt.Sprintf("%d-%d-%d", y, m, d))
	if err != nil {
		// Failed, just return unzeroed
		return p.FromDate
	}
	return zeroed
}

// EndDate returns exactly when this plan ends at
func (p Plan) EndDate() time.Time {
	return p.FromDateZeroHour().Add((time.Hour * 24) * time.Duration(p.DurationDays))
}

// Validate checks the validity of the data inside the Plan object.
// Will return nil if the data is valid.
func (p Plan) Validate() error {
	if p.FromDateZeroHour().Before(time.Now()) {
		return dataerror.ErrBasic("Date cannot be in the past")
	}
	if p.DurationDays == 0 {
		return dataerror.ErrBasic("Duration cannot be zero days")
	}
	if strings.TrimSpace(p.Title) == "" {
		return dataerror.ErrBasic("Title cannot be empty")
	}
	if p.Identifier == "" {
		return dataerror.ErrBasic("No identifier")
	}
	return nil
}

// PlanEntry represents one user's entries in a single plan
type PlanEntry struct {
	gorm.Model
	User            users.User `gorm:"foreignkey:UserID"`
	UserID          uint       `gorm:"not null"`
	PlanID          uint       `gorm:"not null"`
	StartTimeUnix   int64      `gorm:"not null"`
	DurationSeconds int64      `gorm:"not null"`
}

// Validate checks the validity of the data inside the PlanEntry object.
// Will return nil if the data is valid.
func (p PlanEntry) Validate() error {
	if p.StartTimeUnix <= 0 {
		return dataerror.ErrBasic("Why did you think that would work?")
	}
	return nil
}
