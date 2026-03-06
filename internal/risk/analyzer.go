package risk

import (
	"strings"

	"github.com/ofertzadaka/terragrunt-plan-visualizer/internal/plan"
)

// High-risk resource type patterns
var highRiskTypes = []string{
	"aws_iam_role",
	"aws_iam_policy",
	"aws_iam_role_policy",
	"aws_iam_user_policy",
	"aws_iam_group_policy",
	"aws_iam_policy_attachment",
	"aws_iam_role_policy_attachment",
	"aws_security_group",
	"aws_security_group_rule",
	"aws_vpc_security_group_ingress_rule",
	"aws_vpc_security_group_egress_rule",
	"aws_db_instance",
	"aws_rds_cluster",
	"aws_s3_bucket_policy",
	"aws_s3_bucket_public_access_block",
	"aws_kms_key",
	"aws_kms_alias",
	"aws_vpc",
	"aws_subnet",
}

// Medium-risk resource type patterns
var mediumRiskTypes = []string{
	"aws_instance",
	"aws_ecs_service",
	"aws_ecs_task_definition",
	"aws_lambda_function",
	"aws_route_table",
	"aws_route",
	"aws_nat_gateway",
	"aws_autoscaling_group",
	"aws_launch_template",
	"aws_lb",
	"aws_lb_listener",
	"aws_lb_target_group",
}

// Analyze applies risk heuristics to all resource changes in the plan.
func Analyze(p *plan.Plan) {
	for i := range p.ResourceChanges {
		rc := &p.ResourceChanges[i]
		level, reasons := assessRisk(rc)
		rc.RiskLevel = level
		rc.RiskReasons = reasons
	}
	p.Summary = recomputeSummary(p.ResourceChanges)
}

func assessRisk(rc *plan.ResourceChange) (plan.RiskLevel, []string) {
	var reasons []string
	level := plan.RiskLow

	// Any destroy is high risk
	if rc.Action == plan.ActionDelete || rc.Action == plan.ActionReplace ||
		rc.Action == plan.ActionCreateBeforeDelete || rc.Action == plan.ActionDeleteBeforeCreate {
		reasons = append(reasons, "resource will be destroyed")
		level = plan.RiskHigh
	}

	// Check resource type against high-risk list
	if isHighRiskType(rc.Type) {
		if level != plan.RiskHigh {
			level = plan.RiskHigh
		}
		reasons = append(reasons, "sensitive resource type: "+rc.Type)
	} else if isMediumRiskType(rc.Type) {
		if level == plan.RiskLow {
			level = plan.RiskMedium
		}
		reasons = append(reasons, "infrastructure resource type: "+rc.Type)
	}

	// Tag-only changes are low risk
	if rc.Action == plan.ActionUpdate && isTagOnlyChange(rc.Attributes) {
		level = plan.RiskLow
		reasons = []string{"tag-only change"}
	}

	// New resource additions with no destroy are low risk (unless sensitive type)
	if rc.Action == plan.ActionCreate && !isHighRiskType(rc.Type) && !isMediumRiskType(rc.Type) {
		level = plan.RiskLow
		reasons = []string{"new resource addition"}
	}

	if len(reasons) == 0 {
		reasons = []string{"standard change"}
	}

	return level, reasons
}

func isHighRiskType(resourceType string) bool {
	for _, t := range highRiskTypes {
		if strings.EqualFold(resourceType, t) {
			return true
		}
	}
	return false
}

func isMediumRiskType(resourceType string) bool {
	for _, t := range mediumRiskTypes {
		if strings.EqualFold(resourceType, t) {
			return true
		}
	}
	return false
}

func isTagOnlyChange(attrs []plan.AttributeChange) bool {
	if len(attrs) == 0 {
		return false
	}
	for _, attr := range attrs {
		name := strings.ToLower(attr.Name)
		if name != "tags" && name != "tags_all" {
			return false
		}
	}
	return true
}

func recomputeSummary(changes []plan.ResourceChange) plan.PlanSummary {
	s := plan.PlanSummary{TotalChanges: len(changes)}
	for _, c := range changes {
		switch c.Action {
		case plan.ActionCreate:
			s.Adds++
		case plan.ActionUpdate:
			s.Changes++
		case plan.ActionDelete:
			s.Destroys++
		case plan.ActionReplace, plan.ActionCreateBeforeDelete, plan.ActionDeleteBeforeCreate:
			s.Replaces++
		}
		switch c.RiskLevel {
		case plan.RiskHigh:
			s.HighRisk++
		case plan.RiskMedium:
			s.MediumRisk++
		case plan.RiskLow:
			s.LowRisk++
		}
	}
	return s
}
