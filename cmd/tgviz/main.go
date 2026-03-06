package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/ofertzadaka/terragrunt-plan-visualizer/internal/output"
	"github.com/ofertzadaka/terragrunt-plan-visualizer/internal/plan"
	"github.com/ofertzadaka/terragrunt-plan-visualizer/internal/risk"
	"github.com/ofertzadaka/terragrunt-plan-visualizer/internal/server"
	"github.com/spf13/cobra"
)

var (
	port         int
	noBrowser    bool
	planTimeout  time.Duration
	applyTimeout time.Duration
	jsonOnly     bool
	version      = "dev"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "tgviz",
		Short: "Terragrunt plan visualizer with web UI and Claude Code integration",
	}

	rootCmd.PersistentFlags().IntVar(&port, "port", 9090, "Web UI port")
	rootCmd.PersistentFlags().BoolVar(&noBrowser, "no-browser", false, "Don't open browser automatically")
	rootCmd.PersistentFlags().DurationVar(&planTimeout, "plan-timeout", 30*time.Minute, "Plan execution timeout")
	rootCmd.PersistentFlags().DurationVar(&applyTimeout, "apply-timeout", 60*time.Minute, "Apply execution timeout")
	rootCmd.PersistentFlags().BoolVar(&jsonOnly, "json", false, "JSON-only output, no web UI")

	rootCmd.AddCommand(planCmd())
	rootCmd.AddCommand(showCmd())
	rootCmd.AddCommand(applyCmd())
	rootCmd.AddCommand(unlockCmd())
	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func planCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan [dir] [-- tg-args...]",
		Short: "Run terragrunt plan, parse, visualize",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}

			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
			defer cancel()

			extraArgs := cmd.ArgsLenAtDash()
			var tgArgs []string
			if extraArgs >= 0 {
				tgArgs = args[extraArgs:]
			}

			cfg := plan.RunnerConfig{
				WorkingDir:   dir,
				PlanTimeout:  planTimeout,
				ApplyTimeout: applyTimeout,
				ExtraArgs:    tgArgs,
			}

			fmt.Fprintln(os.Stderr, "Running terragrunt plan...")
			jsonData, planFile, err := plan.RunPlan(ctx, cfg)
			if err != nil {
				return handleError(err)
			}

			p, err := plan.ParsePlanJSON(jsonData)
			if err != nil {
				cliErr := &plan.CLIError{Type: "terraform", Message: err.Error()}
				return output.PrintError(cliErr)
			}

			if len(p.ResourceChanges) == 0 {
				fmt.Fprintln(os.Stderr, "No changes. Infrastructure is up to date.")
				return output.PrintNoChanges()
			}

			risk.Analyze(p)
			p.PlanFile = planFile
			p.WorkingDir = dir
			p.Timestamp = time.Now()

			if jsonOnly {
				return output.PrintPlan(p)
			}

			// Start web UI
			srv := server.New(port, p)
			fmt.Fprintf(os.Stderr, "Web UI available at http://localhost:%d\n", port)

			if !noBrowser {
				server.OpenBrowser(fmt.Sprintf("http://localhost:%d", port))
			}

			// Also output JSON for Claude Code
			if err := output.PrintPlan(p); err != nil {
				return err
			}

			return srv.ListenAndServe(ctx)
		},
	}
	return cmd
}

func showCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <plan-json-file>",
		Short: "Visualize an existing plan JSON file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("failed to read plan file: %w", err)
			}

			p, err := plan.ParsePlanJSON(data)
			if err != nil {
				return fmt.Errorf("failed to parse plan JSON: %w", err)
			}

			risk.Analyze(p)
			p.Timestamp = time.Now()

			if jsonOnly {
				return output.PrintPlan(p)
			}

			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
			defer cancel()

			srv := server.New(port, p)
			fmt.Fprintf(os.Stderr, "Web UI available at http://localhost:%d\n", port)

			if !noBrowser {
				server.OpenBrowser(fmt.Sprintf("http://localhost:%d", port))
			}

			return srv.ListenAndServe(ctx)
		},
	}
	return cmd
}

func applyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply [dir]",
		Short: "Apply a previously created plan",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}

			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
			defer cancel()

			cfg := plan.RunnerConfig{
				WorkingDir:   dir,
				ApplyTimeout: applyTimeout,
			}

			planFile := dir + "/tfplan"

			fmt.Fprintln(os.Stderr, "Running terragrunt apply...")
			err := plan.RunApply(ctx, cfg, planFile, func(line string) {
				fmt.Fprintln(os.Stderr, line)
			})

			if err != nil {
				return handleError(err)
			}

			result := &plan.ApplyResult{
				Success:   true,
				Timestamp: time.Now(),
			}
			return output.PrintApplyResult(result)
		},
	}
	return cmd
}

func unlockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unlock <lock-id> [dir]",
		Short: "Force-unlock terraform state",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			lockID := args[0]
			dir := "."
			if len(args) > 1 {
				dir = args[1]
			}

			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
			defer cancel()

			fmt.Fprintf(os.Stderr, "Force-unlocking state (lock ID: %s)...\n", lockID)
			if err := plan.RunUnlock(ctx, dir, lockID); err != nil {
				cliErr := plan.CLIErrorFromErr(err)
				return output.PrintError(cliErr)
			}

			fmt.Fprintln(os.Stderr, "State unlocked successfully.")
			return nil
		},
	}
	return cmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("tgviz %s\n", version)
		},
	}
}

func handleError(err error) error {
	if lockErr, ok := err.(*plan.StateLockError); ok {
		lock := &lockErr.Lock
		cliErr := &lockErr.CLIError
		return output.PrintLockError(cliErr, lock)
	}
	cliErr := plan.CLIErrorFromErr(err)
	return output.PrintError(cliErr)
}
