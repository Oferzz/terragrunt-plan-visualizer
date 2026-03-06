package risk

import (
	"testing"

	"github.com/ofertzadaka/terragrunt-plan-visualizer/internal/plan"
)

func TestAnalyze_DestroyIsHighRisk(t *testing.T) {
	p := &plan.Plan{
		ResourceChanges: []plan.ResourceChange{
			{
				Address: "aws_s3_bucket.example",
				Type:    "aws_s3_bucket",
				Name:    "example",
				Action:  plan.ActionDelete,
			},
		},
	}

	Analyze(p)

	rc := p.ResourceChanges[0]
	if rc.RiskLevel != plan.RiskHigh {
		t.Errorf("expected high risk for delete, got %s", rc.RiskLevel)
	}
	if len(rc.RiskReasons) == 0 {
		t.Error("expected risk reasons for delete")
	}
}

func TestAnalyze_IAMIsHighRisk(t *testing.T) {
	types := []string{
		"aws_iam_role",
		"aws_iam_policy",
		"aws_iam_role_policy",
		"aws_iam_role_policy_attachment",
	}

	for _, resourceType := range types {
		t.Run(resourceType, func(t *testing.T) {
			p := &plan.Plan{
				ResourceChanges: []plan.ResourceChange{
					{
						Address: resourceType + ".test",
						Type:    resourceType,
						Name:    "test",
						Action:  plan.ActionUpdate,
					},
				},
			}

			Analyze(p)

			if p.ResourceChanges[0].RiskLevel != plan.RiskHigh {
				t.Errorf("expected high risk for %s update, got %s", resourceType, p.ResourceChanges[0].RiskLevel)
			}
		})
	}
}

func TestAnalyze_SecurityGroupIsHighRisk(t *testing.T) {
	p := &plan.Plan{
		ResourceChanges: []plan.ResourceChange{
			{
				Address: "aws_security_group.web",
				Type:    "aws_security_group",
				Name:    "web",
				Action:  plan.ActionUpdate,
			},
		},
	}

	Analyze(p)

	if p.ResourceChanges[0].RiskLevel != plan.RiskHigh {
		t.Errorf("expected high risk for security group, got %s", p.ResourceChanges[0].RiskLevel)
	}
}

func TestAnalyze_EC2IsMediumRisk(t *testing.T) {
	p := &plan.Plan{
		ResourceChanges: []plan.ResourceChange{
			{
				Address: "aws_instance.web",
				Type:    "aws_instance",
				Name:    "web",
				Action:  plan.ActionUpdate,
			},
		},
	}

	Analyze(p)

	if p.ResourceChanges[0].RiskLevel != plan.RiskMedium {
		t.Errorf("expected medium risk for EC2 update, got %s", p.ResourceChanges[0].RiskLevel)
	}
}

func TestAnalyze_LambdaIsMediumRisk(t *testing.T) {
	p := &plan.Plan{
		ResourceChanges: []plan.ResourceChange{
			{
				Address: "aws_lambda_function.handler",
				Type:    "aws_lambda_function",
				Name:    "handler",
				Action:  plan.ActionUpdate,
			},
		},
	}

	Analyze(p)

	if p.ResourceChanges[0].RiskLevel != plan.RiskMedium {
		t.Errorf("expected medium risk for Lambda update, got %s", p.ResourceChanges[0].RiskLevel)
	}
}

func TestAnalyze_TagOnlyChangeIsLowRisk(t *testing.T) {
	p := &plan.Plan{
		ResourceChanges: []plan.ResourceChange{
			{
				Address: "aws_instance.web",
				Type:    "aws_instance",
				Name:    "web",
				Action:  plan.ActionUpdate,
				Attributes: []plan.AttributeChange{
					{Name: "tags", OldValue: map[string]interface{}{"env": "dev"}, NewValue: map[string]interface{}{"env": "staging"}},
					{Name: "tags_all", OldValue: map[string]interface{}{"env": "dev"}, NewValue: map[string]interface{}{"env": "staging"}},
				},
			},
		},
	}

	Analyze(p)

	if p.ResourceChanges[0].RiskLevel != plan.RiskLow {
		t.Errorf("expected low risk for tag-only change, got %s", p.ResourceChanges[0].RiskLevel)
	}
}

func TestAnalyze_NewResourceAdditionIsLowRisk(t *testing.T) {
	p := &plan.Plan{
		ResourceChanges: []plan.ResourceChange{
			{
				Address: "aws_cloudwatch_log_group.app",
				Type:    "aws_cloudwatch_log_group",
				Name:    "app",
				Action:  plan.ActionCreate,
			},
		},
	}

	Analyze(p)

	if p.ResourceChanges[0].RiskLevel != plan.RiskLow {
		t.Errorf("expected low risk for new non-sensitive resource, got %s", p.ResourceChanges[0].RiskLevel)
	}
}

func TestAnalyze_NewIAMResourceIsHighRisk(t *testing.T) {
	p := &plan.Plan{
		ResourceChanges: []plan.ResourceChange{
			{
				Address: "aws_iam_role.admin",
				Type:    "aws_iam_role",
				Name:    "admin",
				Action:  plan.ActionCreate,
			},
		},
	}

	Analyze(p)

	if p.ResourceChanges[0].RiskLevel != plan.RiskHigh {
		t.Errorf("expected high risk for new IAM role, got %s", p.ResourceChanges[0].RiskLevel)
	}
}

func TestAnalyze_ReplaceIsHighRisk(t *testing.T) {
	p := &plan.Plan{
		ResourceChanges: []plan.ResourceChange{
			{
				Address: "aws_s3_bucket.data",
				Type:    "aws_s3_bucket",
				Name:    "data",
				Action:  plan.ActionDeleteBeforeCreate,
			},
		},
	}

	Analyze(p)

	if p.ResourceChanges[0].RiskLevel != plan.RiskHigh {
		t.Errorf("expected high risk for replace, got %s", p.ResourceChanges[0].RiskLevel)
	}
}

func TestAnalyze_SummaryRiskCounts(t *testing.T) {
	p := &plan.Plan{
		ResourceChanges: []plan.ResourceChange{
			{Address: "aws_iam_role.a", Type: "aws_iam_role", Action: plan.ActionUpdate},
			{Address: "aws_instance.b", Type: "aws_instance", Action: plan.ActionUpdate},
			{Address: "aws_cloudwatch_log_group.c", Type: "aws_cloudwatch_log_group", Action: plan.ActionCreate},
		},
	}

	Analyze(p)

	if p.Summary.HighRisk != 1 {
		t.Errorf("expected 1 high risk, got %d", p.Summary.HighRisk)
	}
	if p.Summary.MediumRisk != 1 {
		t.Errorf("expected 1 medium risk, got %d", p.Summary.MediumRisk)
	}
	if p.Summary.LowRisk != 1 {
		t.Errorf("expected 1 low risk, got %d", p.Summary.LowRisk)
	}
}

func TestAnalyze_DatabaseIsHighRisk(t *testing.T) {
	types := []string{"aws_db_instance", "aws_rds_cluster"}
	for _, rt := range types {
		t.Run(rt, func(t *testing.T) {
			p := &plan.Plan{
				ResourceChanges: []plan.ResourceChange{
					{Address: rt + ".main", Type: rt, Name: "main", Action: plan.ActionUpdate},
				},
			}
			Analyze(p)
			if p.ResourceChanges[0].RiskLevel != plan.RiskHigh {
				t.Errorf("expected high risk for %s, got %s", rt, p.ResourceChanges[0].RiskLevel)
			}
		})
	}
}

func TestAnalyze_KMSIsHighRisk(t *testing.T) {
	p := &plan.Plan{
		ResourceChanges: []plan.ResourceChange{
			{Address: "aws_kms_key.main", Type: "aws_kms_key", Name: "main", Action: plan.ActionUpdate},
		},
	}
	Analyze(p)
	if p.ResourceChanges[0].RiskLevel != plan.RiskHigh {
		t.Errorf("expected high risk for KMS key, got %s", p.ResourceChanges[0].RiskLevel)
	}
}
