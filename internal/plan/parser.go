package plan

import (
	"encoding/json"
	"fmt"
)

// tfPlan matches the terraform show -json output format.
type tfPlan struct {
	FormatVersion   string             `json:"format_version"`
	TerraformVersion string            `json:"terraform_version"`
	ResourceChanges []tfResourceChange `json:"resource_changes"`
}

type tfResourceChange struct {
	Address      string   `json:"address"`
	Type         string   `json:"type"`
	Name         string   `json:"name"`
	ProviderName string   `json:"provider_name"`
	Change       tfChange `json:"change"`
	ActionReason string   `json:"action_reason,omitempty"`
}

type tfChange struct {
	Actions         []string               `json:"actions"`
	Before          map[string]interface{} `json:"before"`
	After           map[string]interface{} `json:"after"`
	AfterUnknown    map[string]interface{} `json:"after_unknown,omitempty"`
	BeforeSensitive interface{}            `json:"before_sensitive,omitempty"`
	AfterSensitive  interface{}            `json:"after_sensitive,omitempty"`
}

// ParsePlanJSON parses terraform show -json output into our Plan model.
func ParsePlanJSON(data []byte) (*Plan, error) {
	var tf tfPlan
	if err := json.Unmarshal(data, &tf); err != nil {
		return nil, fmt.Errorf("failed to parse terraform plan JSON: %w", err)
	}

	p := &Plan{
		FormatVersion:    tf.FormatVersion,
		TerraformVersion: tf.TerraformVersion,
	}

	for _, rc := range tf.ResourceChanges {
		action := mapActions(rc.Change.Actions)
		if action == ActionNoOp || action == ActionRead {
			continue
		}

		attrs := diffAttributes(rc.Change.Before, rc.Change.After, rc.Change.AfterUnknown)

		change := ResourceChange{
			Address:      rc.Address,
			Type:         rc.Type,
			Name:         rc.Name,
			ProviderName: rc.ProviderName,
			Action:       action,
			ActionReason: rc.ActionReason,
			Attributes:   attrs,
		}

		p.ResourceChanges = append(p.ResourceChanges, change)
	}

	p.Summary = computeSummary(p.ResourceChanges)
	return p, nil
}

// mapActions converts terraform's action array to our Action type.
func mapActions(actions []string) Action {
	if len(actions) == 0 {
		return ActionNoOp
	}
	if len(actions) == 1 {
		switch actions[0] {
		case "create":
			return ActionCreate
		case "delete":
			return ActionDelete
		case "update":
			return ActionUpdate
		case "read":
			return ActionRead
		case "no-op":
			return ActionNoOp
		}
	}
	if len(actions) == 2 {
		if actions[0] == "create" && actions[1] == "delete" {
			return ActionCreateBeforeDelete
		}
		if actions[0] == "delete" && actions[1] == "create" {
			return ActionDeleteBeforeCreate
		}
	}
	return ActionNoOp
}

// diffAttributes computes attribute-level diffs between before and after states.
func diffAttributes(before, after, afterUnknown map[string]interface{}) []AttributeChange {
	var changes []AttributeChange
	seen := make(map[string]bool)

	// Check all keys in after
	for key, newVal := range after {
		seen[key] = true
		oldVal, existed := before[key]

		computed := isComputed(afterUnknown, key)

		if !existed {
			changes = append(changes, AttributeChange{
				Name:     key,
				OldValue: nil,
				NewValue: newVal,
				Computed: computed,
			})
			continue
		}

		if !jsonEqual(oldVal, newVal) {
			changes = append(changes, AttributeChange{
				Name:     key,
				OldValue: oldVal,
				NewValue: newVal,
				Computed: computed,
			})
		}
	}

	// Check for removed keys
	for key, oldVal := range before {
		if !seen[key] {
			changes = append(changes, AttributeChange{
				Name:     key,
				OldValue: oldVal,
				NewValue: nil,
			})
		}
	}

	return changes
}

func isComputed(afterUnknown map[string]interface{}, key string) bool {
	if afterUnknown == nil {
		return false
	}
	val, ok := afterUnknown[key]
	if !ok {
		return false
	}
	b, ok := val.(bool)
	return ok && b
}

// jsonEqual compares two values by marshaling them to JSON.
func jsonEqual(a, b interface{}) bool {
	aj, err1 := json.Marshal(a)
	bj, err2 := json.Marshal(b)
	if err1 != nil || err2 != nil {
		return false
	}
	return string(aj) == string(bj)
}

func computeSummary(changes []ResourceChange) PlanSummary {
	s := PlanSummary{TotalChanges: len(changes)}
	for _, c := range changes {
		switch c.Action {
		case ActionCreate:
			s.Adds++
		case ActionUpdate:
			s.Changes++
		case ActionDelete:
			s.Destroys++
		case ActionReplace, ActionCreateBeforeDelete, ActionDeleteBeforeCreate:
			s.Replaces++
		}
		switch c.RiskLevel {
		case RiskHigh:
			s.HighRisk++
		case RiskMedium:
			s.MediumRisk++
		case RiskLow:
			s.LowRisk++
		}
	}
	return s
}
