package output

import (
	"fmt"
	"io"
	"os"

	"github.com/ofertzadaka/terragrunt-plan-visualizer/internal/plan"
)

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorCyan   = "\033[36m"
	colorDim    = "\033[2m"
	colorBold   = "\033[1m"
)

// PrintFeatureSummary writes the feature analysis summary bar to stderr.
func PrintFeatureSummary(w io.Writer, fc *plan.FeatureContext) {
	if fc == nil {
		return
	}

	if fc.Error != "" {
		fmt.Fprintf(w, " %sFEATURE%s  %s%s%s\n", colorBold, colorReset, colorDim, fc.Error, colorReset)
		return
	}

	fmt.Fprintf(w, " %sFEATURE%s  %s●%d expected%s  %s●%d indirect%s  %s●%d unrelated%s\n",
		colorBold, colorReset,
		colorGreen, fc.ExpectedCount, colorReset,
		colorYellow, fc.IndirectCount, colorReset,
		colorRed, fc.UnrelatedCount, colorReset,
	)
	fmt.Fprintf(w, "  %sbase: %s | %d files changed%s\n",
		colorDim, fc.BaseBranch, len(fc.FilesChanged), colorReset,
	)
}

// PrintResourceWithFeature writes a single resource line with feature relevance indicator.
func PrintResourceWithFeature(w io.Writer, rc *plan.ResourceChange) {
	var featureTag string
	switch rc.FeatureRelevance {
	case plan.RelevanceExpected:
		featureTag = colorGreen + "FEAT" + colorReset
	case plan.RelevanceIndirect:
		featureTag = colorYellow + "IND" + colorReset
	case plan.RelevanceUnrelated:
		featureTag = colorRed + "UNR" + colorReset
	}

	actionSymbol := actionSymbolFor(rc.Action)
	fmt.Fprintf(w, "  %s %s %s[%s]%s\n", actionSymbol, rc.Address, colorDim, featureTag, colorReset)
}

// PrintPlanWithAnalysis outputs the plan summary to stderr with feature and risk info.
func PrintPlanWithAnalysis(p *plan.Plan) {
	w := os.Stderr
	s := p.Summary

	fmt.Fprintf(w, "\n %sPLAN%s  %s+%d%s  %s~%d%s  %s-%d%s  %s!%d%s\n",
		colorBold, colorReset,
		colorGreen, s.Adds, colorReset,
		colorYellow, s.Changes, colorReset,
		colorRed, s.Destroys, colorReset,
		colorCyan, s.Replaces, colorReset,
	)

	fmt.Fprintf(w, " %sRISK%s  %s●%d high%s  %s●%d medium%s  %s●%d low%s\n",
		colorBold, colorReset,
		colorRed, s.HighRisk, colorReset,
		colorYellow, s.MediumRisk, colorReset,
		colorGreen, s.LowRisk, colorReset,
	)

	if p.FeatureContext != nil {
		PrintFeatureSummary(w, p.FeatureContext)
	}

	fmt.Fprintln(w)
}

func actionSymbolFor(action plan.Action) string {
	switch action {
	case plan.ActionCreate:
		return colorGreen + "+" + colorReset
	case plan.ActionUpdate:
		return colorYellow + "~" + colorReset
	case plan.ActionDelete:
		return colorRed + "-" + colorReset
	case plan.ActionReplace, plan.ActionCreateBeforeDelete, plan.ActionDeleteBeforeCreate:
		return colorCyan + "!" + colorReset
	default:
		return " "
	}
}
