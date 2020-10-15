package plan

// CreatePlanRequest is the JSON request object for creating a new plan
type CreatePlanRequest struct {
	Title                  string `json:"title"`
	StartDate              string `json:"start_date"`
	DurationDays           uint   `json:"duration_days"`
	MinAvailabilitySeconds uint   `json:"min_availability_seconds"`
}

// AddEntryRequest is the JSON request object for creating a new entry
type AddEntryRequest struct {
	StartTime       int64 `json:"start_time_unix"`
	DurationSeconds int64 `json:"duration_seconds"`
}
