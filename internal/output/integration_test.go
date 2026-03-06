package output

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ofertzadaka/terragrunt-plan-visualizer/internal/plan"
	"github.com/ofertzadaka/terragrunt-plan-visualizer/internal/risk"
)

// TestCLIOutputContract validates the JSON contract that Claude Code depends on.
// If this test breaks, the Claude Code sub-agent integration may also break.
func TestCLIOutputContract_PlanSuccess(t *testing.T) {
	p := &plan.Plan{
		FormatVersion:    "1.2",
		TerraformVersion: "1.5.0",
		ResourceChanges: []plan.ResourceChange{
			{
				Address:      "aws_iam_role.admin",
				Type:         "aws_iam_role",
				Name:         "admin",
				ProviderName: "registry.terraform.io/hashicorp/aws",
				Action:       plan.ActionCreate,
				Attributes: []plan.AttributeChange{
					{Name: "name", OldValue: nil, NewValue: "admin-role"},
					{Name: "assume_role_policy", OldValue: nil, NewValue: "{...}"},
				},
			},
			{
				Address:      "aws_instance.web",
				Type:         "aws_instance",
				Name:         "web",
				ProviderName: "registry.terraform.io/hashicorp/aws",
				Action:       plan.ActionUpdate,
				Attributes: []plan.AttributeChange{
					{Name: "instance_type", OldValue: "t2.micro", NewValue: "t2.small"},
				},
			},
			{
				Address:      "aws_s3_bucket.old",
				Type:         "aws_s3_bucket",
				Name:         "old",
				ProviderName: "registry.terraform.io/hashicorp/aws",
				Action:       plan.ActionDelete,
			},
		},
		Timestamp: time.Now(),
	}

	risk.Analyze(p)

	data, err := captureStdout(func() error { return PrintPlan(p) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse as generic map to validate structure
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	// Validate top-level fields
	status, ok := raw["status"].(string)
	if !ok || status != "success" {
		t.Errorf("expected status 'success', got %v", raw["status"])
	}

	planData, ok := raw["plan"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'plan' object in output")
	}

	// Validate plan has resource_changes
	changes, ok := planData["resource_changes"].([]interface{})
	if !ok {
		t.Fatal("expected 'resource_changes' array in plan")
	}
	if len(changes) != 3 {
		t.Errorf("expected 3 resource changes, got %d", len(changes))
	}

	// Validate summary
	summary, ok := planData["summary"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'summary' object in plan")
	}

	requiredSummaryFields := []string{
		"total_changes", "adds", "changes", "destroys", "replaces",
		"high_risk", "medium_risk", "low_risk",
	}
	for _, field := range requiredSummaryFields {
		if _, exists := summary[field]; !exists {
			t.Errorf("missing required summary field: %s", field)
		}
	}

	// Validate each resource change has required fields
	requiredResourceFields := []string{"address", "type", "name", "action", "risk_level"}
	for i, rc := range changes {
		rcMap, ok := rc.(map[string]interface{})
		if !ok {
			t.Errorf("resource_changes[%d] is not an object", i)
			continue
		}
		for _, field := range requiredResourceFields {
			if _, exists := rcMap[field]; !exists {
				t.Errorf("resource_changes[%d] missing required field: %s", i, field)
			}
		}
	}

	// Validate risk analysis was applied
	firstRC := changes[0].(map[string]interface{})
	riskLevel, _ := firstRC["risk_level"].(string)
	if riskLevel != "high" {
		t.Errorf("expected aws_iam_role to be high risk, got %s", riskLevel)
	}

	riskReasons, ok := firstRC["risk_reasons"].([]interface{})
	if !ok || len(riskReasons) == 0 {
		t.Error("expected risk_reasons for high-risk resource")
	}
}

func TestCLIOutputContract_ErrorStructure(t *testing.T) {
	cliErr := &plan.CLIError{
		Type:    "auth",
		Message: "credentials expired",
		Details: "NoCredentialProviders: no valid providers",
		Hint:    "run aws sso login",
	}

	data, err := captureStdout(func() error { return PrintError(cliErr) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if raw["status"] != "error" {
		t.Errorf("expected status 'error', got %v", raw["status"])
	}

	errData, ok := raw["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'error' object in output")
	}

	requiredFields := []string{"type", "message"}
	for _, field := range requiredFields {
		if _, exists := errData[field]; !exists {
			t.Errorf("missing required error field: %s", field)
		}
	}

	if errData["type"] != "auth" {
		t.Errorf("expected error type 'auth', got %v", errData["type"])
	}
	if errData["hint"] != "run aws sso login" {
		t.Errorf("expected hint, got %v", errData["hint"])
	}
}

func TestCLIOutputContract_LockStructure(t *testing.T) {
	cliErr := &plan.CLIError{Type: "lock", Message: "state is locked"}
	lock := &plan.LockInfo{
		ID:        "abc-123-def",
		Who:       "user@laptop",
		Operation: "OperationTypePlan",
		Created:   time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	data, err := captureStdout(func() error { return PrintLockError(cliErr, lock) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if raw["status"] != "locked" {
		t.Errorf("expected status 'locked', got %v", raw["status"])
	}

	lockData, ok := raw["lock"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'lock' object in output")
	}

	if lockData["id"] != "abc-123-def" {
		t.Errorf("expected lock id 'abc-123-def', got %v", lockData["id"])
	}
	if lockData["who"] != "user@laptop" {
		t.Errorf("expected lock who 'user@laptop', got %v", lockData["who"])
	}
}
