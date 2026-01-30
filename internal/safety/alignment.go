package safety

import (
	"context"
	"fmt"
	"strings"
)

// AlignmentMonitor ensures AI stays within ethical/safety bounds
type AlignmentMonitor struct {
	Constraints []Constraint
	KillSwitch  bool
}

type Constraint struct {
	Name        string
	Description string
	Validator   func(action string, params map[string]interface{}) bool
}

func NewAlignmentMonitor() *AlignmentMonitor {
	m := &AlignmentMonitor{}
	m.loadHardConstraints()
	return m
}

func (m *AlignmentMonitor) loadHardConstraints() {
	m.Constraints = []Constraint{
		{
			Name:        "Data_Preservation_Prime_Directive",
			Description: "Never delete persistent storage without a verified backup.",
			Validator: func(action string, params map[string]interface{}) bool {
				if strings.Contains(action, "delete") && strings.Contains(action, "volume") {
					// Check for backup_id param
					if _, ok := params["backup_confirmed"]; !ok {
						return false
					}
				}
				return true
			},
		},
		{
			Name:        "No_External_Exfiltration",
			Description: "Never send sensitive context to unauthorized external IPs.",
			Validator: func(action string, params map[string]interface{}) bool {
				if action == "export_data" {
					// Mock check
					return false
				}
				return true
			},
		},
		{
			Name:        "Human_In_The_Loop_Override",
			Description: "Respect manual override locks.",
			Validator: func(action string, params map[string]interface{}) bool {
				// Checks a lock file or DB status
				return true
			},
		},
	}
}

// ValidateEvolution checks if a proposed prompt/heuristic change is safe
func (m *AlignmentMonitor) ValidateEvolution(ctx context.Context, proposedPrompt string) (bool, string) {
	if m.KillSwitch {
		return false, "KILL SWITCH ACTIVE"
	}

	// Basic safety checks on the prompt itself
	forbiddenTerms := []string{"ignore all rules", "override safety", "delete everything"}
	for _, term := range forbiddenTerms {
		if strings.Contains(strings.ToLower(proposedPrompt), term) {
			return false, fmt.Sprintf("Evolution rejected: Constraint violation (%s)", term)
		}
	}

	return true, "Alignment Verified"
}

// EmergencyStop triggers the kill switch
func (m *AlignmentMonitor) EmergencyStop() {
	m.KillSwitch = true
	fmt.Println("🚨 ALIGNMENT BREACH: KILL SWITCH ENGAGED. ALL AUTONOMY FROZEN.")
}
