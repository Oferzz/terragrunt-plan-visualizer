package plan

import "time"

// Action represents the type of change for a resource.
type Action string

const (
	ActionCreate            Action = "create"
	ActionUpdate            Action = "update"
	ActionDelete            Action = "delete"
	ActionReplace           Action = "replace"
	ActionRead              Action = "read"
	ActionNoOp              Action = "no-op"
	ActionCreateBeforeDelete Action = "create-before-delete"
	ActionDeleteBeforeCreate Action = "delete-before-create"
)

// FeatureRelevance indicates how a resource change relates to the current git diff.
type FeatureRelevance string

const (
	RelevanceExpected  FeatureRelevance = "expected"
	RelevanceIndirect  FeatureRelevance = "indirect"
	RelevanceUnrelated FeatureRelevance = "unrelated"
)

// FeatureContext holds the result of git-diff-based feature analysis.
type FeatureContext struct {
	BaseBranch      string   `json:"base_branch"`
	FilesChanged    []string `json:"files_changed"`
	ResourcesInDiff []string `json:"resources_in_diff"`
	ModulesInDiff   []string `json:"modules_in_diff"`
	ExpectedCount   int      `json:"expected_count"`
	IndirectCount   int      `json:"indirect_count"`
	UnrelatedCount  int      `json:"unrelated_count"`
	Error           string   `json:"error,omitempty"`
}

// RiskLevel represents the severity of a change.
type RiskLevel string

const (
	RiskHigh   RiskLevel = "high"
	RiskMedium RiskLevel = "medium"
	RiskLow    RiskLevel = "low"
)

// AttributeChange represents a single attribute diff within a resource.
type AttributeChange struct {
	Name     string      `json:"name"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
	Computed bool        `json:"computed,omitempty"`
}

// ResourceChange represents a single resource change from the plan.
type ResourceChange struct {
	Address      string            `json:"address"`
	Type         string            `json:"type"`
	Name         string            `json:"name"`
	ProviderName string            `json:"provider_name"`
	Action       Action            `json:"action"`
	ActionReason string            `json:"action_reason,omitempty"`
	Attributes   []AttributeChange `json:"attributes,omitempty"`
	RiskLevel        RiskLevel        `json:"risk_level"`
	RiskReasons      []string         `json:"risk_reasons,omitempty"`
	FeatureRelevance FeatureRelevance `json:"feature_relevance,omitempty"`
	FeatureReason    string           `json:"feature_reason,omitempty"`
}

// PlanSummary provides high-level counts of changes.
type PlanSummary struct {
	TotalChanges int `json:"total_changes"`
	Adds         int `json:"adds"`
	Changes      int `json:"changes"`
	Destroys     int `json:"destroys"`
	Replaces     int `json:"replaces"`
	HighRisk     int `json:"high_risk"`
	MediumRisk   int `json:"medium_risk"`
	LowRisk      int `json:"low_risk"`
}

// Plan represents a fully parsed and analyzed terraform plan.
type Plan struct {
	FormatVersion   string           `json:"format_version"`
	TerraformVersion string          `json:"terraform_version"`
	ResourceChanges []ResourceChange `json:"resource_changes"`
	Summary         PlanSummary      `json:"summary"`
	FeatureContext  *FeatureContext  `json:"feature_context,omitempty"`
	PlanFile        string           `json:"plan_file,omitempty"`
	WorkingDir      string           `json:"working_dir,omitempty"`
	Timestamp       time.Time        `json:"timestamp"`
}

// ApplyResult represents the outcome of an apply operation.
type ApplyResult struct {
	Success           bool             `json:"success"`
	ResourcesCreated  int              `json:"resources_created"`
	ResourcesModified int              `json:"resources_modified"`
	ResourcesDestroyed int             `json:"resources_destroyed"`
	Errors            []ApplyError     `json:"errors,omitempty"`
	Timestamp         time.Time        `json:"timestamp"`
}

// ApplyError represents a single error during apply.
type ApplyError struct {
	Resource string `json:"resource"`
	Message  string `json:"message"`
}

// LockInfo contains details about a terraform state lock.
type LockInfo struct {
	ID        string    `json:"id"`
	Operation string    `json:"operation"`
	Who       string    `json:"who"`
	Created   time.Time `json:"created"`
}

// CLIOutput is the top-level JSON structure output by tgviz for Claude Code consumption.
type CLIOutput struct {
	Status  string          `json:"status"` // "success", "error", "locked", "no_changes"
	Plan    *Plan           `json:"plan,omitempty"`
	Apply   *ApplyResult    `json:"apply,omitempty"`
	Lock    *LockInfo       `json:"lock,omitempty"`
	Error   *CLIError       `json:"error,omitempty"`
	Feature *FeatureContext `json:"feature,omitempty"`
}

// CLIError represents a structured error from the CLI.
type CLIError struct {
	Type    string `json:"type"` // "auth", "network", "terraform", "timeout", "lock", "unknown"
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Hint    string `json:"hint,omitempty"`
}

func (e *CLIError) Error() string {
	return e.Message
}
