package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ofertzadaka/terragrunt-plan-visualizer/internal/plan"
)

// PrintPlan outputs the plan as structured JSON to stdout.
func PrintPlan(p *plan.Plan) error {
	out := plan.CLIOutput{
		Status:  "success",
		Plan:    p,
		Feature: p.FeatureContext,
	}
	return printJSON(out)
}

// PrintNoChanges outputs a no-changes result to stdout.
func PrintNoChanges() error {
	out := plan.CLIOutput{
		Status: "no_changes",
	}
	return printJSON(out)
}

// PrintApplyResult outputs the apply result as structured JSON to stdout.
func PrintApplyResult(result *plan.ApplyResult) error {
	status := "success"
	if !result.Success {
		status = "error"
	}
	out := plan.CLIOutput{
		Status: status,
		Apply:  result,
	}
	return printJSON(out)
}

// PrintError outputs a structured error to stdout.
func PrintError(cliErr *plan.CLIError) error {
	out := plan.CLIOutput{
		Status: "error",
		Error:  cliErr,
	}
	return printJSON(out)
}

// PrintLockError outputs a lock error with lock details to stdout.
func PrintLockError(cliErr *plan.CLIError, lock *plan.LockInfo) error {
	out := plan.CLIOutput{
		Status: "locked",
		Lock:   lock,
		Error:  cliErr,
	}
	return printJSON(out)
}

func printJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("failed to write JSON output: %w", err)
	}
	return nil
}
