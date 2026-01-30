package engine

// Opportunity represents a detected optimization chance
type Opportunity struct {
	ID          string
	ResourceID  string
	Description string
	Savings     float64
	Confidence  float64
	ActionType  string // e.g., "resize", "stop", "switch_region"
}

// PredictSavings returns an estimated dollar amount
func (o *Opportunity) PredictSavings() float64 {
	return o.Savings
}
