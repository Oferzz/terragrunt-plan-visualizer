package risk

import (
	"fmt"

	"github.com/ofertzadaka/terragrunt-plan-visualizer/internal/plan"
)

// GenerateFindings produces risk findings including feature-aware drift warnings.
func GenerateFindings(p *plan.Plan) []string {
	var findings []string

	// Standard risk findings
	for _, rc := range p.ResourceChanges {
		if rc.RiskLevel == plan.RiskHigh {
			findings = append(findings, fmt.Sprintf("HIGH RISK: %s (%s) - %s", rc.Address, rc.Type, rc.Action))
		}
	}

	// Feature-aware drift warnings
	if p.FeatureContext != nil && p.FeatureContext.Error == "" {
		if p.FeatureContext.UnrelatedCount > 0 {
			findings = append(findings, fmt.Sprintf(
				"DRIFT: %d resource change(s) are NOT related to your code changes",
				p.FeatureContext.UnrelatedCount,
			))
		}

		// Flag dangerous unrelated destroys
		for _, rc := range p.ResourceChanges {
			if rc.FeatureRelevance == plan.RelevanceUnrelated &&
				(rc.Action == plan.ActionDelete || rc.Action == plan.ActionReplace ||
					rc.Action == plan.ActionCreateBeforeDelete || rc.Action == plan.ActionDeleteBeforeCreate) {
				findings = append(findings, fmt.Sprintf(
					"DANGER - UNRELATED DESTROY: %s (%s) is being destroyed but is NOT part of your feature changes",
					rc.Address, rc.Type,
				))
			}
		}
	}

	return findings
}

// GenerateRecommendations produces recommendations based on risk and feature analysis.
func GenerateRecommendations(p *plan.Plan) []string {
	var recs []string

	if p.Summary.HighRisk > 0 {
		recs = append(recs, "Review all high-risk changes carefully before applying")
	}

	if p.Summary.Destroys > 0 {
		recs = append(recs, "Verify destroyed resources are intentional and backed up if needed")
	}

	// Feature-aware recommendations
	if p.FeatureContext != nil && p.FeatureContext.Error == "" {
		if p.FeatureContext.UnrelatedCount > 0 {
			recs = append(recs, fmt.Sprintf(
				"Investigate %d unrelated change(s) before applying - these may indicate infrastructure drift",
				p.FeatureContext.UnrelatedCount,
			))
		}
	}

	if len(recs) == 0 {
		recs = append(recs, "Changes appear safe to apply")
	}

	return recs
}
