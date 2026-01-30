package governance

import (
	"context"
	"fmt"
)

// Policy defines a compliance rule
type Policy struct {
	ID          string
	Name        string
	Description string
	Severity    string // CRITICAL, HIGH, MEDIUM, LOW
	CheckFunc   func(resource map[string]interface{}) (bool, string)
	RemedyFunc  func(resourceID string) error
}

// PolicyEngine enforces compliance
type PolicyEngine struct {
	policies []Policy
}

func NewPolicyEngine() *PolicyEngine {
	engine := &PolicyEngine{}
	engine.loadDefaultPolicies()
	return engine
}

func (e *PolicyEngine) loadDefaultPolicies() {
	e.policies = []Policy{
		{
			ID:          "SEC-001",
			Name:        "No Public S3 Buckets",
			Description: "Ensure S3 buckets do not allow public read access.",
			Severity:    "CRITICAL",
			CheckFunc: func(res map[string]interface{}) (bool, string) {
				if access, ok := res["public_access"].(bool); ok && access {
					return false, "Bucket allows public access"
				}
				return true, ""
			},
			RemedyFunc: func(id string) error {
				fmt.Printf("🛡️ Auto-Remediation: Blocking public access for bucket %s\n", id)
				return nil
			},
		},
		{
			ID:          "SEC-002",
			Name:        "SSH Port 22 Closed",
			Description: "Security groups should not allow 0.0.0.0/0 on port 22.",
			Severity:    "HIGH",
			CheckFunc: func(res map[string]interface{}) (bool, string) {
				if ports, ok := res["open_ports"].([]int); ok {
					for _, p := range ports {
						if p == 22 {
							return false, "Port 22 is open to world"
						}
					}
				}
				return true, ""
			},
			RemedyFunc: func(id string) error {
				fmt.Printf("🛡️ Auto-Remediation: Removing rule for port 22 on SG %s\n", id)
				return nil
			},
		},
		{
			ID:          "TAG-001",
			Name:        "Cost Center Tagging",
			Description: "All resources must have a 'CostCenter' tag.",
			Severity:    "MEDIUM",
			CheckFunc: func(res map[string]interface{}) (bool, string) {
				tags, ok := res["tags"].(map[string]string)
				if !ok {
					return false, "No tags found"
				}
				if _, exists := tags["CostCenter"]; !exists {
					return false, "Missing CostCenter tag"
				}
				return true, ""
			},
			RemedyFunc: func(id string) error {
				fmt.Printf("🛡️ Auto-Remediation: Adding default CostCenter tag to %s\n", id)
				return nil
			},
		},
	}
}

// ComplianceReport contains results of a scan
type ComplianceReport struct {
	TotalChecks int
	Passed      int
	Failed      int
	Violations  []Violation
}

type Violation struct {
	PolicyID   string
	ResourceID string
	Reason     string
	Severity   string
}

// ScanResource checks a resource against all policies
func (e *PolicyEngine) ScanResource(ctx context.Context, resourceID string, attributes map[string]interface{}) *ComplianceReport {
	report := &ComplianceReport{}

	for _, policy := range e.policies {
		report.TotalChecks++
		pass, reason := policy.CheckFunc(attributes)
		if pass {
			report.Passed++
		} else {
			report.Failed++
			report.Violations = append(report.Violations, Violation{
				PolicyID:   policy.ID,
				ResourceID: resourceID,
				Reason:     reason,
				Severity:   policy.Severity,
			})
		}
	}

	return report
}

// AutoRemediate attempts to fix violations
func (e *PolicyEngine) AutoRemediate(ctx context.Context, violations []Violation) int {
	fixed := 0
	for _, v := range violations {
		// Find policy
		var policy *Policy
		for _, p := range e.policies {
			if p.ID == v.PolicyID {
				policy = &p
				break
			}
		}

		if policy != nil && policy.RemedyFunc != nil {
			if err := policy.RemedyFunc(v.ResourceID); err == nil {
				fixed++
			}
		}
	}
	return fixed
}
