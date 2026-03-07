package output

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/ofertzadaka/terragrunt-plan-visualizer/internal/plan"
)

func captureStdout(fn func() error) ([]byte, error) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := fn()
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.Bytes(), err
}

func TestPrintPlan(t *testing.T) {
	p := &plan.Plan{
		FormatVersion:    "1.2",
		TerraformVersion: "1.5.0",
		ResourceChanges: []plan.ResourceChange{
			{Address: "aws_s3_bucket.test", Type: "aws_s3_bucket", Action: plan.ActionCreate, RiskLevel: plan.RiskLow},
		},
		Summary:   plan.PlanSummary{TotalChanges: 1, Adds: 1, LowRisk: 1},
		Timestamp: time.Now(),
	}

	data, err := captureStdout(func() error { return PrintPlan(p) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out plan.CLIOutput
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("failed to parse output JSON: %v", err)
	}
	if out.Status != "success" {
		t.Errorf("expected status success, got %s", out.Status)
	}
	if out.Plan == nil {
		t.Fatal("expected plan in output")
	}
	if len(out.Plan.ResourceChanges) != 1 {
		t.Errorf("expected 1 resource change, got %d", len(out.Plan.ResourceChanges))
	}
}

func TestPrintNoChanges(t *testing.T) {
	data, err := captureStdout(func() error { return PrintNoChanges() })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out plan.CLIOutput
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("failed to parse output JSON: %v", err)
	}
	if out.Status != "no_changes" {
		t.Errorf("expected status no_changes, got %s", out.Status)
	}
}

func TestPrintError(t *testing.T) {
	cliErr := &plan.CLIError{
		Type:    "auth",
		Message: "authentication failed",
		Hint:    "run aws sso login",
	}

	data, err := captureStdout(func() error { return PrintError(cliErr) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out plan.CLIOutput
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("failed to parse output JSON: %v", err)
	}
	if out.Status != "error" {
		t.Errorf("expected status error, got %s", out.Status)
	}
	if out.Error == nil {
		t.Fatal("expected error in output")
	}
	if out.Error.Type != "auth" {
		t.Errorf("expected error type auth, got %s", out.Error.Type)
	}
}

func TestPrintLockError(t *testing.T) {
	cliErr := &plan.CLIError{Type: "lock", Message: "state is locked"}
	lock := &plan.LockInfo{ID: "abc123", Who: "user@host", Operation: "OperationTypePlan"}

	data, err := captureStdout(func() error { return PrintLockError(cliErr, lock) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out plan.CLIOutput
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("failed to parse output JSON: %v", err)
	}
	if out.Status != "locked" {
		t.Errorf("expected status locked, got %s", out.Status)
	}
	if out.Lock == nil {
		t.Fatal("expected lock info in output")
	}
	if out.Lock.ID != "abc123" {
		t.Errorf("expected lock ID abc123, got %s", out.Lock.ID)
	}
}

func TestPrintPlanWithFeatureContext(t *testing.T) {
	p := &plan.Plan{
		FormatVersion:    "1.2",
		TerraformVersion: "1.5.0",
		ResourceChanges: []plan.ResourceChange{
			{
				Address: "aws_s3_bucket.test", Type: "aws_s3_bucket", Action: plan.ActionCreate,
				RiskLevel: plan.RiskLow, FeatureRelevance: plan.RelevanceExpected, FeatureReason: "direct match",
			},
		},
		Summary:   plan.PlanSummary{TotalChanges: 1, Adds: 1, LowRisk: 1},
		Timestamp: time.Now(),
		FeatureContext: &plan.FeatureContext{
			BaseBranch:      "main",
			FilesChanged:    []string{"main.tf"},
			ResourcesInDiff: []string{"aws_s3_bucket.test"},
			ExpectedCount:   1,
		},
	}

	data, err := captureStdout(func() error { return PrintPlan(p) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out plan.CLIOutput
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("failed to parse output JSON: %v", err)
	}
	if out.Feature == nil {
		t.Fatal("expected feature context in output")
	}
	if out.Feature.BaseBranch != "main" {
		t.Errorf("expected base branch 'main', got %s", out.Feature.BaseBranch)
	}
	if out.Feature.ExpectedCount != 1 {
		t.Errorf("expected expected_count=1, got %d", out.Feature.ExpectedCount)
	}
	if out.Plan.ResourceChanges[0].FeatureRelevance != plan.RelevanceExpected {
		t.Errorf("expected feature_relevance='expected', got %s", out.Plan.ResourceChanges[0].FeatureRelevance)
	}
}

func TestPrintPlanWithoutFeatureContext(t *testing.T) {
	p := &plan.Plan{
		FormatVersion:    "1.2",
		TerraformVersion: "1.5.0",
		ResourceChanges: []plan.ResourceChange{
			{Address: "aws_s3_bucket.test", Type: "aws_s3_bucket", Action: plan.ActionCreate, RiskLevel: plan.RiskLow},
		},
		Summary:   plan.PlanSummary{TotalChanges: 1, Adds: 1, LowRisk: 1},
		Timestamp: time.Now(),
	}

	data, err := captureStdout(func() error { return PrintPlan(p) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to parse output JSON: %v", err)
	}
	if _, exists := raw["feature"]; exists {
		t.Error("expected no 'feature' key when feature context is nil")
	}

	// Also verify feature_relevance is omitted from resource changes
	planData := raw["plan"].(map[string]interface{})
	changes := planData["resource_changes"].([]interface{})
	rc := changes[0].(map[string]interface{})
	if _, exists := rc["feature_relevance"]; exists {
		t.Error("expected no 'feature_relevance' when not set")
	}
}

func TestPrintApplyResult(t *testing.T) {
	result := &plan.ApplyResult{
		Success:          true,
		ResourcesCreated: 3,
		Timestamp:        time.Now(),
	}

	data, err := captureStdout(func() error { return PrintApplyResult(result) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out plan.CLIOutput
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("failed to parse output JSON: %v", err)
	}
	if out.Status != "success" {
		t.Errorf("expected status success, got %s", out.Status)
	}
	if out.Apply == nil {
		t.Fatal("expected apply result in output")
	}
	if out.Apply.ResourcesCreated != 3 {
		t.Errorf("expected 3 created, got %d", out.Apply.ResourcesCreated)
	}
}
