package feature

import (
	"fmt"
	"strings"

	"github.com/ofertzadaka/terragrunt-plan-visualizer/internal/plan"
)

// Correlate classifies each ResourceChange in the plan based on the git diff.
// It modifies ResourceChanges in-place and returns a FeatureContext summary.
func Correlate(p *plan.Plan, diff *DiffResult) *plan.FeatureContext {
	ctx := &plan.FeatureContext{
		BaseBranch:      diff.BaseBranch,
		FilesChanged:    diff.FilesChanged,
		ResourcesInDiff: diff.ResourceIdentifiers(),
		ModulesInDiff:   diff.ModuleNames(),
	}

	if diff.Error != "" {
		ctx.Error = diff.Error
	}

	// Build lookup sets
	resourceSet := buildResourceSet(diff.Resources)
	typeSet := buildTypeSet(diff.Resources)
	moduleSet := buildModuleSet(diff.Modules)

	// First pass: classify each resource
	for i := range p.ResourceChanges {
		rc := &p.ResourceChanges[i]
		relevance, reason := classify(rc, resourceSet, typeSet, moduleSet)
		rc.FeatureRelevance = relevance
		rc.FeatureReason = reason
	}

	// Second pass: tag propagation detection
	// If a resource is update-only with tag-only changes and another resource
	// of the same provider is "expected", mark it as "indirect"
	expectedProviders := map[string]bool{}
	for _, rc := range p.ResourceChanges {
		if rc.FeatureRelevance == plan.RelevanceExpected {
			provider := extractProvider(rc.Type)
			expectedProviders[provider] = true
		}
	}

	for i := range p.ResourceChanges {
		rc := &p.ResourceChanges[i]
		if rc.FeatureRelevance == plan.RelevanceUnrelated && isTagOnlyChange(rc.Attributes) {
			provider := extractProvider(rc.Type)
			if expectedProviders[provider] {
				rc.FeatureRelevance = plan.RelevanceIndirect
				rc.FeatureReason = "tag-only change, likely propagated from related resources"
			}
		}
	}

	// Count results
	for _, rc := range p.ResourceChanges {
		switch rc.FeatureRelevance {
		case plan.RelevanceExpected:
			ctx.ExpectedCount++
		case plan.RelevanceIndirect:
			ctx.IndirectCount++
		default:
			ctx.UnrelatedCount++
		}
	}

	return ctx
}

func classify(rc *plan.ResourceChange, resourceSet map[string]bool, typeSet map[string]bool, moduleSet map[string]bool) (plan.FeatureRelevance, string) {
	addr := stripAddressIndices(rc.Address)
	resType, resName := extractResourceIdentifier(rc.Address)
	resKey := resType + "." + resName

	// 1. Direct match: type.name from plan matches a DiffResource
	if resourceSet[resKey] {
		return plan.RelevanceExpected, fmt.Sprintf("direct match: %s found in git diff", resKey)
	}

	// 2. Module match: plan address contains module.X where X is in diff modules
	moduleName := extractModuleName(rc.Address)
	if moduleName != "" && moduleSet[moduleName] {
		return plan.RelevanceExpected, fmt.Sprintf("module match: module.%s found in git diff", moduleName)
	}

	// Also check for nested module matches with stripped address
	for part := range moduleSet {
		if strings.Contains(addr, "module."+part+".") {
			return plan.RelevanceExpected, fmt.Sprintf("module match: module.%s found in git diff", part)
		}
	}

	// 3. Type-level match: same resource type exists in diff but different name
	if typeSet[resType] {
		return plan.RelevanceIndirect, fmt.Sprintf("same type %s found in diff with different name", resType)
	}

	// 4. No match
	return plan.RelevanceUnrelated, "no connection to git diff found"
}

func buildResourceSet(resources []DiffResource) map[string]bool {
	set := make(map[string]bool, len(resources))
	for _, r := range resources {
		set[r.Type+"."+r.Name] = true
	}
	return set
}

func buildTypeSet(resources []DiffResource) map[string]bool {
	set := make(map[string]bool, len(resources))
	for _, r := range resources {
		set[r.Type] = true
	}
	return set
}

func buildModuleSet(modules []DiffModule) map[string]bool {
	set := make(map[string]bool, len(modules))
	for _, m := range modules {
		set[m.Name] = true
	}
	return set
}

func extractProvider(resourceType string) string {
	// aws_s3_bucket -> aws, google_compute_instance -> google
	parts := strings.SplitN(resourceType, "_", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return resourceType
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
