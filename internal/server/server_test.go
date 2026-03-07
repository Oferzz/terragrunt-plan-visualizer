package server

import (
	"bytes"
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

func TestHandleApply_MethodNotAllowed(t *testing.T) {
	s := New(0, testPlan())

	req := httptest.NewRequest("GET", "/api/apply", nil)
	w := httptest.NewRecorder()

	s.handleApply(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleApply_NoApplyFunc(t *testing.T) {
	s := New(0, testPlan())

	req := httptest.NewRequest("POST", "/api/apply", nil)
	w := httptest.NewRecorder()

	s.handleApply(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

func TestHandleAIAnalysis_GetEmpty(t *testing.T) {
	s := New(0, testPlan())

	req := httptest.NewRequest("GET", "/api/ai-analysis", nil)
	w := httptest.NewRecorder()

	s.handleAIAnalysis(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestHandlePlan_WithFeatureContext(t *testing.T) {
	p := testPlan()
	p.FeatureContext = &plan.FeatureContext{
		BaseBranch:      "main",
		FilesChanged:    []string{"main.tf"},
		ResourcesInDiff: []string{"aws_s3_bucket.test"},
		ExpectedCount:   1,
	}
	p.ResourceChanges[0].FeatureRelevance = plan.RelevanceExpected
	p.ResourceChanges[0].FeatureReason = "direct match"

	s := New(0, p)

	req := httptest.NewRequest("GET", "/api/plan", nil)
	w := httptest.NewRecorder()

	s.handlePlan(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var result plan.Plan
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result.FeatureContext == nil {
		t.Fatal("expected feature_context in response")
	}
	if result.FeatureContext.BaseBranch != "main" {
		t.Errorf("expected base_branch 'main', got %s", result.FeatureContext.BaseBranch)
	}
	if result.ResourceChanges[0].FeatureRelevance != plan.RelevanceExpected {
		t.Errorf("expected feature_relevance 'expected', got %s", result.ResourceChanges[0].FeatureRelevance)
	}
}

func TestHandleAIAnalysis_PostAndGet(t *testing.T) {
	s := New(0, testPlan())

	analysis := AIAnalysisData{
		Findings:        []string{"Open security group detected"},
		RiskSummary:     "Medium risk overall",
		Recommendations: []string{"Restrict ingress rules"},
	}
	body, _ := json.Marshal(analysis)

	// POST analysis
	postReq := httptest.NewRequest("POST", "/api/ai-analysis", bytes.NewReader(body))
	postW := httptest.NewRecorder()
	s.handleAIAnalysis(postW, postReq)

	if postW.Code != http.StatusOK {
		t.Errorf("POST expected 200, got %d", postW.Code)
	}

	// GET it back
	getReq := httptest.NewRequest("GET", "/api/ai-analysis", nil)
	getW := httptest.NewRecorder()
	s.handleAIAnalysis(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Errorf("GET expected 200, got %d", getW.Code)
	}

	var result AIAnalysisData
	if err := json.Unmarshal(getW.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(result.Findings) != 1 || result.Findings[0] != "Open security group detected" {
		t.Errorf("unexpected findings: %v", result.Findings)
	}
	if result.RiskSummary != "Medium risk overall" {
		t.Errorf("unexpected risk summary: %s", result.RiskSummary)
	}
}
