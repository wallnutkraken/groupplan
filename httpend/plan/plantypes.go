package plan

// CreatePlanRequest is the JSON request object for creating a new plan
type CreatePlanRequest struct {
	Title        string `json:"title"`
	StartDate    string `json:"start_date"`
	DurationDays uint   `json:"duration_days"`
}
