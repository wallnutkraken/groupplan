// Package planman is the plan manager, it's responsible for plan operations with the data layer
package planman

import (
	"fmt"
	"time"

	"github.com/wallnutkraken/groupplan/groupdata/dataerror"

	"github.com/wallnutkraken/groupplan/groupdata/plans"
	"github.com/wallnutkraken/groupplan/groupdata/users"
	"github.com/wallnutkraken/groupplan/secid"
	"github.com/wallnutkraken/groupplan/userman"
)

// Planner is responsible for plan operations with the data layer
type Planner struct {
	data PlanData
}

// PlanData is the interface for what methods the plan persistency layer should provide PlanMan
type PlanData interface {
	CreatePlan(plan *plans.Plan) error
	GetPlan(identifier string) (plan plans.Plan, err error)
	DeletePlan(plan plans.Plan) error
	AddEntry(plan *plans.Plan, user users.User, availFrom, duration int64) (plans.PlanEntry, error)
	GetPlansByUser(user users.User) ([]plans.Plan, error)
	DeleteEntry(entryID uint) error
	GetEntry(entryID uint) (entry plans.PlanEntry, err error)
	GetEntriesOnPlanByUser(planID string, user users.User) ([]plans.PlanEntry, error)
}

// New creates a new instance of the PlanMan Planner
func New(db PlanData) Planner {
	return Planner{
		data: db,
	}
}

// NewPlan creates a new plan, owned by the given User
// The caller should call errors.Is on the error returned from this function to check if it's
// a dataerror.ValidationErrors error
func (p Planner) NewPlan(title string, fromDate time.Time, durationDays uint, owner users.User) (GroupPlan, error) {
	identifier, err := secid.String(16)
	if err != nil {
		return GroupPlan{}, fmt.Errorf("failed creating secure identifier: %w", err)
	}
	plan := plans.Plan{
		Owner:        owner,
		OwnerID:      owner.ID,
		Identifier:   identifier,
		Title:        title,
		FromDate:     fromDate,
		DurationDays: durationDays,
	}
	if err := p.data.CreatePlan(&plan); err != nil {
		return GroupPlan{}, fmt.Errorf("failed creating the plan in the database: %w", err)
	}

	// Take the newly-saved plan object and turn it into the local GroupPlan one
	groupPlan := GroupPlan{}
	groupPlan.FillFromDataType(plan)

	// And return it with no error
	return groupPlan, nil
}

// GetPlans returns a list of all the user's plans
func (p Planner) GetPlans(user users.User) ([]GroupPlan, error) {
	plans, err := p.data.GetPlansByUser(user)
	if err != nil {
		// Just return this one without wrapping, it'd just be redundant info here
		return nil, err
	}
	// Change the data types into GroupPlans
	groupPlans := make([]GroupPlan, len(plans))
	for index, plan := range plans {
		current := GroupPlan{}
		current.FillFromDataType(plan)
		groupPlans[index] = current
	}

	return groupPlans, nil
}

// GetEntriesOnPlanByUser gets a list of availability entries for a given user on the
// specified plan
func (p Planner) GetEntriesOnPlanByUser(planIdentifier string, user users.User) ([]PlanEntry, error) {
	entries, err := p.data.GetEntriesOnPlanByUser(planIdentifier, user)
	if err != nil {
		return nil, err
	}
	// Convert the entries into PlanEntry
	converted := make([]PlanEntry, len(entries))
	for index, entry := range entries {
		conv := PlanEntry{}
		conv.FillFromDataType(entry)
		converted[index] = conv
	}
	return converted, nil
}

// AddEntry creates a new entry for availability for a plan, identified by the given identifier.
func (p Planner) AddEntry(planIdentifier string, user users.User, startAtUnix, duration int64) (PlanEntry, error) {
	// Get the plan based on identifier
	plan, err := p.data.GetPlan(planIdentifier)
	if err != nil {
		return PlanEntry{}, fmt.Errorf("no plan: %w", err)
	}
	// Check that the duration is longer than the plan's minimum availability
	if plan.MinimumAvailabilitySeconds > uint(duration) {
		return PlanEntry{}, dataerror.ErrBasic(fmt.Sprintf("Entry duration cannot be shorter than the plan's (%d)", plan.MinimumAvailabilitySeconds))
	}

	createdEntry, err := p.data.AddEntry(&plan, user, startAtUnix, duration)
	if err != nil {
		return PlanEntry{}, fmt.Errorf("failed saving entry: %w", err)
	}

	// Now just convert createdEntry to PlanEntry
	finalEntry := PlanEntry{}
	finalEntry.FillFromDataType(createdEntry)

	return finalEntry, nil
}

// DeletePlan deletes a plan with the given identifier if the owner of the plan is the given user
func (p Planner) DeletePlan(identifier string, user users.User) error {
	// Get the plan to check the owner
	plan, err := p.data.GetPlan(identifier)
	if err != nil {
		return fmt.Errorf("could not get plan [%s]: %w", identifier, err)
	}
	if plan.OwnerID != user.ID {
		return dataerror.ErrUnauthorized("you are not the owner of this plan")
	}

	// This user is the owner of the plan, delete it
	return p.data.DeletePlan(plan)
}

// DeleteEntry deletes an availability entry inside a plan with the given entry ID, if the user
// is the owner of the entry
func (p Planner) DeleteEntry(entryID uint, user users.User) error {
	// Get the entry first to check if the user is the owner
	entry, err := p.data.GetEntry(entryID)
	if err != nil {
		return fmt.Errorf("could not get entry: %w", err)
	}
	if entry.UserID != user.ID {
		return dataerror.ErrUnauthorized("you are not the owner of this entry")
	}

	// User is the owner, delete it
	return p.data.DeleteEntry(entry.ID)
}

// GetPlan gets a plan from the data layer with the given identifier
func (p Planner) GetPlan(identifier string) (GroupPlan, error) {
	plan, err := p.data.GetPlan(identifier)
	if err != nil {
		return GroupPlan{}, fmt.Errorf("could not get plan with identifier [%s]: %w", identifier, err)
	}
	// Convert plan to GroupPlan
	groupPlan := GroupPlan{}
	groupPlan.FillFromDataType(plan)

	return groupPlan, nil
}

// GroupPlan represents a single plan
type GroupPlan struct {
	Owner               userman.User `json:"owner"`
	Identifier          string       `json:"identifier"`
	Title               string       `json:"title"`
	FromDate            time.Time    `json:"from_date"`
	DurationDays        uint         `json:"duration_days"`
	MinAvailabilitySecs uint         `json:"min_availability_seconds"`
	Entries             []PlanEntry  `json:"entries"`
}

// PlanEntry contains the specifics of a single plan entry
type PlanEntry struct {
	EntryID         uint         `json:"entry_id"`
	User            userman.User `json:"user"`
	StartAtUnix     int64        `json:"start_at_unix"`
	DurationSeconds int64        `json:"duration_seconds"`
}

// FillFromDataType fills the GroupPlan object from the provided database type
func (g *GroupPlan) FillFromDataType(plan plans.Plan) {
	g.Owner = userman.User{
		DisplayName: plan.Owner.DisplayName,
		AvatarURL:   plan.Owner.ProfilePictureURL,
	}
	g.Identifier = plan.Identifier
	g.Title = plan.Title
	g.FromDate = plan.FromDate
	g.DurationDays = plan.DurationDays
	g.Entries = make([]PlanEntry, len(plan.Entries))
	g.MinAvailabilitySecs = plan.MinimumAvailabilitySeconds
	for index, entry := range plan.Entries {
		fill := PlanEntry{}
		fill.FillFromDataType(entry)
		g.Entries[index] = fill
	}
}

// FillFromDataType fills the PlanEntry object from the provided database type
func (p *PlanEntry) FillFromDataType(entry plans.PlanEntry) {
	p.EntryID = entry.ID
	p.User = userman.User{
		DisplayName: entry.User.DisplayName,
		AvatarURL:   entry.User.ProfilePictureURL,
	}
	p.StartAtUnix = entry.StartTimeUnix
	p.DurationSeconds = entry.DurationSeconds
}
