package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ofertzadaka/terragrunt-plan-visualizer/internal/plan"
)

func testPlan() *plan.Plan {
	return &plan.Plan{
		FormatVersion:    "1.2",
		TerraformVersion: "1.5.0",
		ResourceChanges: []plan.ResourceChange{
			{
				Address:   "aws_s3_bucket.test",
				Type:      "aws_s3_bucket",
				Name:      "test",
				Action:    plan.ActionCreate,
				RiskLevel: plan.RiskLow,
			},
		},
		Summary: plan.PlanSummary{
			TotalChanges: 1,
			Adds:         1,
			LowRisk:      1,
		},
		Timestamp: time.Now(),
	}
}

func TestHandlePlan(t *testing.T) {
	s := New(0, testPlan())

	req := httptest.NewRequest("GET", "/api/plan", nil)
	w := httptest.NewRecorder()

	s.handlePlan(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}

	var p plan.Plan
	if err := json.Unmarshal(w.Body.Bytes(), &p); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}
	if len(p.ResourceChanges) != 1 {
		t.Errorf("expected 1 resource change, got %d", len(p.ResourceChanges))
	}
	if p.ResourceChanges[0].Address != "aws_s3_bucket.test" {
		t.Errorf("expected aws_s3_bucket.test, got %s", p.ResourceChanges[0].Address)
	}
}

func TestHandleStatus(t *testing.T) {
	s := New(0, testPlan())

	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()

	s.handleStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if result["status"] != "ready" {
		t.Errorf("expected status ready, got %s", result["status"])
	}
}

func TestHandleIndex(t *testing.T) {
	s := New(0, testPlan())

	mux := http.NewServeMux()
	s.registerRoutes(mux)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("expected text/html content-type, got %s", ct)
	}
	body := w.Body.String()
	if len(body) == 0 {
		t.Error("expected non-empty HTML body")
	}
}
