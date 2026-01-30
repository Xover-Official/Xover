package cloud

import (
	"encoding/json"
	"fmt"
)

type TerraformPlan struct {
	ResourceChanges []ResourceChange `json:"resource_changes"`
}

type ResourceChange struct {
	Address string `json:"address"`
	Type    string `json:"type"`
	Change  Change `json:"change"`
}

type Change struct {
	Actions []string               `json:"actions"`
	Before  map[string]interface{} `json:"before"`
	After   map[string]interface{} `json:"after"`
}

type TerraformParser struct{}

func (p *TerraformParser) ParsePlan(jsonData []byte) ([]*ResourceV2, error) {
	var plan TerraformPlan
	if err := json.Unmarshal(jsonData, &plan); err != nil {
		return nil, fmt.Errorf("failed to unmarshal terraform plan: %w", err)
	}

	var resources []*ResourceV2
	for _, rc := range plan.ResourceChanges {
		// Only focus on resources being updated or created
		if len(rc.Change.Actions) == 0 || rc.Change.Actions[0] == "no-op" {
			continue
		}

		// Map Terraform resource types to Atlas types
		resourceType := "unknown"
		switch rc.Type {
		case "aws_db_instance":
			resourceType = "rds"
		case "aws_instance":
			resourceType = "ec2"
		}

		resources = append(resources, &ResourceV2{
			ID:       rc.Address,
			Type:     resourceType,
			Provider: "terraform",
			Region:   "unknown",
			State:    "planned",
			// Additional fields would be populated by the actual cloud provider
		})
	}

	return resources, nil
}
