package feature

import (
	"testing"

	"github.com/ofertzadaka/terragrunt-plan-visualizer/internal/plan"
)

func TestCorrelate_DirectMatch(t *testing.T) {
	p := &plan.Plan{
		ResourceChanges: []plan.ResourceChange{
			{Address: "aws_s3_bucket.my_bucket", Type: "aws_s3_bucket", Name: "my_bucket", Action: plan.ActionCreate},
		},
	}
	diff := &DiffResult{
		Resources: []DiffResource{
			{Type: "aws_s3_bucket", Name: "my_bucket", Action: "added"},
		},
	}

	ctx := Correlate(p, diff)

	if p.ResourceChanges[0].FeatureRelevance != plan.RelevanceExpected {
		t.Errorf("expected 'expected', got %s", p.ResourceChanges[0].FeatureRelevance)
	}
	if ctx.ExpectedCount != 1 {
		t.Errorf("expected ExpectedCount=1, got %d", ctx.ExpectedCount)
	}
}

func TestCorrelate_ModuleMatch(t *testing.T) {
	p := &plan.Plan{
		ResourceChanges: []plan.ResourceChange{
			{Address: "module.vpc.aws_subnet.public", Type: "aws_subnet", Name: "public", Action: plan.ActionCreate},
		},
	}
	diff := &DiffResult{
		Modules: []DiffModule{
			{Name: "vpc", Action: "added"},
		},
	}

	_ = Correlate(p, diff)

	if p.ResourceChanges[0].FeatureRelevance != plan.RelevanceExpected {
		t.Errorf("expected 'expected', got %s", p.ResourceChanges[0].FeatureRelevance)
	}
	if p.ResourceChanges[0].FeatureReason == "" {
		t.Error("expected feature reason to be set")
	}
}

func TestCorrelate_ForEachIndex(t *testing.T) {
	p := &plan.Plan{
		ResourceChanges: []plan.ResourceChange{
			{Address: `aws_subnet.public["us-east-1a"]`, Type: "aws_subnet", Name: "public", Action: plan.ActionCreate},
		},
	}
	diff := &DiffResult{
		Resources: []DiffResource{
			{Type: "aws_subnet", Name: "public", Action: "added"},
		},
	}

	ctx := Correlate(p, diff)

	if p.ResourceChanges[0].FeatureRelevance != plan.RelevanceExpected {
		t.Errorf("expected 'expected', got %s", p.ResourceChanges[0].FeatureRelevance)
	}
	if ctx.ExpectedCount != 1 {
		t.Errorf("expected ExpectedCount=1, got %d", ctx.ExpectedCount)
	}
}

func TestCorrelate_Unrelated(t *testing.T) {
	p := &plan.Plan{
		ResourceChanges: []plan.ResourceChange{
			{Address: "aws_lambda_function.cron", Type: "aws_lambda_function", Name: "cron", Action: plan.ActionUpdate},
		},
	}
	diff := &DiffResult{
		Resources: []DiffResource{
			{Type: "aws_s3_bucket", Name: "data", Action: "added"},
		},
	}

	ctx := Correlate(p, diff)

	if p.ResourceChanges[0].FeatureRelevance != plan.RelevanceUnrelated {
		t.Errorf("expected 'unrelated', got %s", p.ResourceChanges[0].FeatureRelevance)
	}
	if ctx.UnrelatedCount != 1 {
		t.Errorf("expected UnrelatedCount=1, got %d", ctx.UnrelatedCount)
	}
}

func TestCorrelate_IndirectSameType(t *testing.T) {
	p := &plan.Plan{
		ResourceChanges: []plan.ResourceChange{
			{Address: "aws_s3_bucket.other", Type: "aws_s3_bucket", Name: "other", Action: plan.ActionUpdate},
		},
	}
	diff := &DiffResult{
		Resources: []DiffResource{
			{Type: "aws_s3_bucket", Name: "primary", Action: "modified"},
		},
	}

	Correlate(p, diff)

	if p.ResourceChanges[0].FeatureRelevance != plan.RelevanceIndirect {
		t.Errorf("expected 'indirect', got %s", p.ResourceChanges[0].FeatureRelevance)
	}
}

func TestCorrelate_TagPropagation(t *testing.T) {
	p := &plan.Plan{
		ResourceChanges: []plan.ResourceChange{
			{
				Address: "aws_s3_bucket.main", Type: "aws_s3_bucket", Name: "main", Action: plan.ActionCreate,
			},
			{
				Address: "aws_s3_bucket.tagged", Type: "aws_s3_bucket", Name: "tagged", Action: plan.ActionUpdate,
				Attributes: []plan.AttributeChange{
					{Name: "tags", OldValue: map[string]interface{}{"env": "dev"}, NewValue: map[string]interface{}{"env": "prod"}},
				},
			},
		},
	}
	diff := &DiffResult{
		Resources: []DiffResource{
			{Type: "aws_s3_bucket", Name: "main", Action: "added"},
		},
	}

	Correlate(p, diff)

	if p.ResourceChanges[0].FeatureRelevance != plan.RelevanceExpected {
		t.Errorf("expected main bucket to be 'expected', got %s", p.ResourceChanges[0].FeatureRelevance)
	}
	if p.ResourceChanges[1].FeatureRelevance != plan.RelevanceIndirect {
		t.Errorf("expected tagged bucket to be 'indirect', got %s", p.ResourceChanges[1].FeatureRelevance)
	}
}

func TestCorrelate_EmptyDiff(t *testing.T) {
	p := &plan.Plan{
		ResourceChanges: []plan.ResourceChange{
			{Address: "aws_instance.web", Type: "aws_instance", Name: "web", Action: plan.ActionUpdate},
			{Address: "aws_s3_bucket.data", Type: "aws_s3_bucket", Name: "data", Action: plan.ActionCreate},
		},
	}
	diff := &DiffResult{}

	ctx := Correlate(p, diff)

	if ctx.UnrelatedCount != 2 {
		t.Errorf("expected UnrelatedCount=2, got %d", ctx.UnrelatedCount)
	}
	for _, rc := range p.ResourceChanges {
		if rc.FeatureRelevance != plan.RelevanceUnrelated {
			t.Errorf("expected 'unrelated' for %s, got %s", rc.Address, rc.FeatureRelevance)
		}
	}
}

func TestCorrelate_FeatureContextCounts(t *testing.T) {
	p := &plan.Plan{
		ResourceChanges: []plan.ResourceChange{
			{Address: "aws_s3_bucket.main", Type: "aws_s3_bucket", Name: "main", Action: plan.ActionCreate},
			{Address: "aws_s3_bucket.other", Type: "aws_s3_bucket", Name: "other", Action: plan.ActionUpdate},
			{Address: "aws_lambda_function.cron", Type: "aws_lambda_function", Name: "cron", Action: plan.ActionUpdate},
		},
	}
	diff := &DiffResult{
		BaseBranch:   "main",
		FilesChanged: []string{"s3.tf"},
		Resources: []DiffResource{
			{Type: "aws_s3_bucket", Name: "main", Action: "added"},
		},
	}

	ctx := Correlate(p, diff)

	if ctx.ExpectedCount != 1 {
		t.Errorf("expected ExpectedCount=1, got %d", ctx.ExpectedCount)
	}
	if ctx.IndirectCount != 1 {
		t.Errorf("expected IndirectCount=1, got %d", ctx.IndirectCount)
	}
	if ctx.UnrelatedCount != 1 {
		t.Errorf("expected UnrelatedCount=1, got %d", ctx.UnrelatedCount)
	}
	if ctx.BaseBranch != "main" {
		t.Errorf("expected BaseBranch='main', got %s", ctx.BaseBranch)
	}
	if len(ctx.FilesChanged) != 1 {
		t.Errorf("expected 1 file changed, got %d", len(ctx.FilesChanged))
	}
	if len(ctx.ResourcesInDiff) != 1 {
		t.Errorf("expected 1 resource in diff, got %d", len(ctx.ResourcesInDiff))
	}
}

func TestCorrelate_ModuleWithIndex(t *testing.T) {
	p := &plan.Plan{
		ResourceChanges: []plan.ResourceChange{
			{Address: `module.vpc[0].aws_subnet.private`, Type: "aws_subnet", Name: "private", Action: plan.ActionCreate},
		},
	}
	diff := &DiffResult{
		Modules: []DiffModule{
			{Name: "vpc", Action: "added"},
		},
	}

	Correlate(p, diff)

	if p.ResourceChanges[0].FeatureRelevance != plan.RelevanceExpected {
		t.Errorf("expected 'expected', got %s", p.ResourceChanges[0].FeatureRelevance)
	}
}
