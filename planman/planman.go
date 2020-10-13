// Package planman is the plan manager, it's responsible for plan operations with the data layer
package planman

import (
	"time"

	"github.com/wallnutkraken/groupplan/groupdata/plans"
	"github.com/wallnutkraken/groupplan/groupdata/users"
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
	AddEntry(plan *plans.Plan, user users.User, availFrom, availTo int64) (plans.PlanEntry, error)
}

// New creates a new instance of the PlanMan Planner
func New(db PlanData) Planner {
	return Planner{
		data: db,
	}
}

// GroupPlan represents a single plan
type GroupPlan struct {
	Owner      userman.User `json:"owner"`
	Identifier string       `json:"identifier"`
	Title      string       `json:"title"`
	FromDate   time.Time    `json:"from_date"`
	ToDate     time.Time    `json:"to_date"`
	Entries    []PlanEntry  `json:"entry"`
}

// PlanEntry contains the specifics of a single plan entry
type PlanEntry struct {
	EntryID     uint         `json:"entry_id"`
	User        userman.User `json:"user"`
	StartAtUnix int64        `json:"start_at_unix"`
	EndAtUnix   int64        `json:"end_at_unix"`
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
	g.ToDate = plan.ToDate
	g.Entries = make([]PlanEntry, len(plan.Entries))
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
	p.EndAtUnix = entry.EndTimeUnix
}
