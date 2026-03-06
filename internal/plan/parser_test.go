package plan

import (
	"encoding/json"
	"testing"
)

func TestParsePlanJSON_BasicCreate(t *testing.T) {
	input := `{
		"format_version": "1.2",
		"terraform_version": "1.5.0",
		"resource_changes": [
			{
				"address": "aws_s3_bucket.example",
				"type": "aws_s3_bucket",
				"name": "example",
				"provider_name": "registry.terraform.io/hashicorp/aws",
				"change": {
					"actions": ["create"],
					"before": null,
					"after": {
						"bucket": "my-bucket",
						"tags": {"env": "dev"}
					}
				}
			}
		]
	}`

	p, err := parsePlanJSON(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.FormatVersion != "1.2" {
		t.Errorf("expected format_version 1.2, got %s", p.FormatVersion)
	}
	if p.TerraformVersion != "1.5.0" {
		t.Errorf("expected terraform_version 1.5.0, got %s", p.TerraformVersion)
	}
	if len(p.ResourceChanges) != 1 {
		t.Fatalf("expected 1 resource change, got %d", len(p.ResourceChanges))
	}

	rc := p.ResourceChanges[0]
	if rc.Address != "aws_s3_bucket.example" {
		t.Errorf("expected address aws_s3_bucket.example, got %s", rc.Address)
	}
	if rc.Action != ActionCreate {
		t.Errorf("expected action create, got %s", rc.Action)
	}
	if rc.Type != "aws_s3_bucket" {
		t.Errorf("expected type aws_s3_bucket, got %s", rc.Type)
	}

	if p.Summary.Adds != 1 {
		t.Errorf("expected 1 add, got %d", p.Summary.Adds)
	}
	if p.Summary.TotalChanges != 1 {
		t.Errorf("expected 1 total change, got %d", p.Summary.TotalChanges)
	}
}

func TestParsePlanJSON_UpdateWithDiff(t *testing.T) {
	input := `{
		"format_version": "1.2",
		"terraform_version": "1.5.0",
		"resource_changes": [
			{
				"address": "aws_instance.web",
				"type": "aws_instance",
				"name": "web",
				"provider_name": "registry.terraform.io/hashicorp/aws",
				"change": {
					"actions": ["update"],
					"before": {
						"instance_type": "t2.micro",
						"tags": {"env": "dev"}
					},
					"after": {
						"instance_type": "t2.small",
						"tags": {"env": "staging"}
					}
				}
			}
		]
	}`

	p, err := parsePlanJSON(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(p.ResourceChanges) != 1 {
		t.Fatalf("expected 1 resource change, got %d", len(p.ResourceChanges))
	}

	rc := p.ResourceChanges[0]
	if rc.Action != ActionUpdate {
		t.Errorf("expected action update, got %s", rc.Action)
	}

	// Check attribute diffs
	found := make(map[string]bool)
	for _, attr := range rc.Attributes {
		found[attr.Name] = true
		if attr.Name == "instance_type" {
			if attr.OldValue != "t2.micro" {
				t.Errorf("expected old instance_type t2.micro, got %v", attr.OldValue)
			}
			if attr.NewValue != "t2.small" {
				t.Errorf("expected new instance_type t2.small, got %v", attr.NewValue)
			}
		}
	}
	if !found["instance_type"] {
		t.Error("expected instance_type in attribute changes")
	}
	if !found["tags"] {
		t.Error("expected tags in attribute changes")
	}

	if p.Summary.Changes != 1 {
		t.Errorf("expected 1 change, got %d", p.Summary.Changes)
	}
}

func TestParsePlanJSON_Delete(t *testing.T) {
	input := `{
		"format_version": "1.2",
		"terraform_version": "1.5.0",
		"resource_changes": [
			{
				"address": "aws_vpc.main",
				"type": "aws_vpc",
				"name": "main",
				"provider_name": "registry.terraform.io/hashicorp/aws",
				"change": {
					"actions": ["delete"],
					"before": {"cidr_block": "10.0.0.0/16"},
					"after": null
				}
			}
		]
	}`

	p, err := parsePlanJSON(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(p.ResourceChanges) != 1 {
		t.Fatalf("expected 1 resource change, got %d", len(p.ResourceChanges))
	}

	rc := p.ResourceChanges[0]
	if rc.Action != ActionDelete {
		t.Errorf("expected action delete, got %s", rc.Action)
	}
	if p.Summary.Destroys != 1 {
		t.Errorf("expected 1 destroy, got %d", p.Summary.Destroys)
	}
}

func TestParsePlanJSON_Replace(t *testing.T) {
	input := `{
		"format_version": "1.2",
		"terraform_version": "1.5.0",
		"resource_changes": [
			{
				"address": "aws_instance.web",
				"type": "aws_instance",
				"name": "web",
				"provider_name": "registry.terraform.io/hashicorp/aws",
				"change": {
					"actions": ["delete", "create"],
					"before": {"ami": "ami-old"},
					"after": {"ami": "ami-new"}
				},
				"action_reason": "replace_because_cannot_update"
			}
		]
	}`

	p, err := parsePlanJSON(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rc := p.ResourceChanges[0]
	if rc.Action != ActionDeleteBeforeCreate {
		t.Errorf("expected action delete-before-create, got %s", rc.Action)
	}
	if rc.ActionReason != "replace_because_cannot_update" {
		t.Errorf("expected action_reason, got %s", rc.ActionReason)
	}
	if p.Summary.Replaces != 1 {
		t.Errorf("expected 1 replace, got %d", p.Summary.Replaces)
	}
}

func TestParsePlanJSON_SkipsNoOp(t *testing.T) {
	input := `{
		"format_version": "1.2",
		"terraform_version": "1.5.0",
		"resource_changes": [
			{
				"address": "aws_s3_bucket.existing",
				"type": "aws_s3_bucket",
				"name": "existing",
				"provider_name": "registry.terraform.io/hashicorp/aws",
				"change": {
					"actions": ["no-op"],
					"before": {"bucket": "existing"},
					"after": {"bucket": "existing"}
				}
			},
			{
				"address": "aws_s3_bucket.new",
				"type": "aws_s3_bucket",
				"name": "new",
				"provider_name": "registry.terraform.io/hashicorp/aws",
				"change": {
					"actions": ["create"],
					"before": null,
					"after": {"bucket": "new-bucket"}
				}
			}
		]
	}`

	p, err := parsePlanJSON(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(p.ResourceChanges) != 1 {
		t.Fatalf("expected 1 resource change (no-op filtered), got %d", len(p.ResourceChanges))
	}
	if p.ResourceChanges[0].Address != "aws_s3_bucket.new" {
		t.Errorf("expected only the create resource, got %s", p.ResourceChanges[0].Address)
	}
}

func TestParsePlanJSON_MultipleResources(t *testing.T) {
	input := `{
		"format_version": "1.2",
		"terraform_version": "1.5.0",
		"resource_changes": [
			{
				"address": "aws_s3_bucket.a",
				"type": "aws_s3_bucket",
				"name": "a",
				"provider_name": "registry.terraform.io/hashicorp/aws",
				"change": {"actions": ["create"], "before": null, "after": {"bucket": "a"}}
			},
			{
				"address": "aws_instance.b",
				"type": "aws_instance",
				"name": "b",
				"provider_name": "registry.terraform.io/hashicorp/aws",
				"change": {"actions": ["update"], "before": {"tags": {}}, "after": {"tags": {"env": "prod"}}}
			},
			{
				"address": "aws_vpc.c",
				"type": "aws_vpc",
				"name": "c",
				"provider_name": "registry.terraform.io/hashicorp/aws",
				"change": {"actions": ["delete"], "before": {"cidr_block": "10.0.0.0/16"}, "after": null}
			}
		]
	}`

	p, err := parsePlanJSON(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(p.ResourceChanges) != 3 {
		t.Fatalf("expected 3 resource changes, got %d", len(p.ResourceChanges))
	}
	if p.Summary.Adds != 1 {
		t.Errorf("expected 1 add, got %d", p.Summary.Adds)
	}
	if p.Summary.Changes != 1 {
		t.Errorf("expected 1 change, got %d", p.Summary.Changes)
	}
	if p.Summary.Destroys != 1 {
		t.Errorf("expected 1 destroy, got %d", p.Summary.Destroys)
	}
	if p.Summary.TotalChanges != 3 {
		t.Errorf("expected 3 total changes, got %d", p.Summary.TotalChanges)
	}
}

func TestParsePlanJSON_ComputedAttribute(t *testing.T) {
	input := `{
		"format_version": "1.2",
		"terraform_version": "1.5.0",
		"resource_changes": [
			{
				"address": "aws_instance.web",
				"type": "aws_instance",
				"name": "web",
				"provider_name": "registry.terraform.io/hashicorp/aws",
				"change": {
					"actions": ["create"],
					"before": null,
					"after": {"ami": "ami-123"},
					"after_unknown": {"id": true, "arn": true}
				}
			}
		]
	}`

	p, err := parsePlanJSON(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rc := p.ResourceChanges[0]
	for _, attr := range rc.Attributes {
		if attr.Name == "ami" && attr.Computed {
			t.Error("ami should not be marked as computed")
		}
	}
}

func TestParsePlanJSON_InvalidJSON(t *testing.T) {
	_, err := ParsePlanJSON([]byte(`{invalid json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParsePlanJSON_EmptyResourceChanges(t *testing.T) {
	input := `{
		"format_version": "1.2",
		"terraform_version": "1.5.0",
		"resource_changes": []
	}`

	p, err := parsePlanJSON(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.ResourceChanges) != 0 {
		t.Errorf("expected 0 resource changes, got %d", len(p.ResourceChanges))
	}
	if p.Summary.TotalChanges != 0 {
		t.Errorf("expected 0 total changes, got %d", p.Summary.TotalChanges)
	}
}

func TestMapActions(t *testing.T) {
	tests := []struct {
		name     string
		actions  []string
		expected Action
	}{
		{"create", []string{"create"}, ActionCreate},
		{"delete", []string{"delete"}, ActionDelete},
		{"update", []string{"update"}, ActionUpdate},
		{"read", []string{"read"}, ActionRead},
		{"no-op", []string{"no-op"}, ActionNoOp},
		{"create-before-delete", []string{"create", "delete"}, ActionCreateBeforeDelete},
		{"delete-before-create", []string{"delete", "create"}, ActionDeleteBeforeCreate},
		{"empty", []string{}, ActionNoOp},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapActions(tt.actions)
			if got != tt.expected {
				t.Errorf("mapActions(%v) = %s, want %s", tt.actions, got, tt.expected)
			}
		})
	}
}

func TestDiffAttributes_NewAttributes(t *testing.T) {
	before := map[string]interface{}(nil)
	after := map[string]interface{}{
		"bucket": "my-bucket",
		"tags":   map[string]interface{}{"env": "dev"},
	}

	attrs := diffAttributes(before, after, nil)
	if len(attrs) != 2 {
		t.Fatalf("expected 2 attribute changes, got %d", len(attrs))
	}
}

func TestDiffAttributes_RemovedAttributes(t *testing.T) {
	before := map[string]interface{}{
		"bucket": "my-bucket",
		"tags":   map[string]interface{}{"env": "dev"},
	}
	after := map[string]interface{}{}

	attrs := diffAttributes(before, after, nil)
	if len(attrs) != 2 {
		t.Fatalf("expected 2 attribute changes, got %d", len(attrs))
	}
	for _, attr := range attrs {
		if attr.NewValue != nil {
			t.Errorf("expected nil new value for removed attr %s", attr.Name)
		}
	}
}

func TestDiffAttributes_NoChange(t *testing.T) {
	before := map[string]interface{}{"bucket": "same"}
	after := map[string]interface{}{"bucket": "same"}

	attrs := diffAttributes(before, after, nil)
	if len(attrs) != 0 {
		t.Errorf("expected 0 attribute changes for identical values, got %d", len(attrs))
	}
}

func TestParsePlanJSON_RoundTrip(t *testing.T) {
	input := `{
		"format_version": "1.2",
		"terraform_version": "1.5.0",
		"resource_changes": [
			{
				"address": "aws_s3_bucket.example",
				"type": "aws_s3_bucket",
				"name": "example",
				"provider_name": "registry.terraform.io/hashicorp/aws",
				"change": {
					"actions": ["create"],
					"before": null,
					"after": {"bucket": "my-bucket"}
				}
			}
		]
	}`

	p, err := parsePlanJSON(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Marshal and verify it's valid JSON
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("failed to marshal plan: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty JSON output")
	}
}

// Helper to parse plan JSON in tests
func parsePlanJSON(t *testing.T, input string) (*Plan, error) {
	t.Helper()
	return ParsePlanJSON([]byte(input))
}
