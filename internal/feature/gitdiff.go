package feature

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// DiffResult holds the parsed output of a git diff.
type DiffResult struct {
	BaseBranch   string
	FilesChanged []string
	Resources    []DiffResource
	Modules      []DiffModule
	Variables    []string
	Outputs      []string
	Error        string
}

// DiffResource represents a resource block found in the diff.
type DiffResource struct {
	Type     string
	Name     string
	FilePath string
	Action   string // "added", "modified", "removed"
}

// DiffModule represents a module block found in the diff.
type DiffModule struct {
	Name     string
	Source   string
	FilePath string
	Action   string // "added", "modified", "removed"
}

var (
	resourceRe = regexp.MustCompile(`resource\s+"([^"]+)"\s+"([^"]+)"`)
	moduleRe   = regexp.MustCompile(`module\s+"([^"]+)"`)
	dataRe     = regexp.MustCompile(`data\s+"([^"]+)"\s+"([^"]+)"`)
	indexRe    = regexp.MustCompile(`\[[^\]]*\]`)
)

// AnalyzeGitDiff runs git diff against the base branch and parses the result.
func AnalyzeGitDiff(workingDir, baseBranch string) *DiffResult {
	gitRoot, err := findGitRoot(workingDir)
	if err != nil {
		return &DiffResult{Error: fmt.Sprintf("not a git repository: %v", err)}
	}

	diffOutput, err := getGitDiff(gitRoot, baseBranch)
	if err != nil {
		return &DiffResult{Error: fmt.Sprintf("git diff failed: %v", err)}
	}

	result := parseDiff(diffOutput)
	result.BaseBranch = baseBranch
	return result
}

// findGitRoot walks up from workingDir to find the .git directory.
func findGitRoot(workingDir string) (string, error) {
	out, err := runCommand("git", "-C", workingDir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// getGitDiff runs git diff against the base branch, filtering to .tf and .hcl files.
func getGitDiff(gitRoot, baseBranch string) (string, error) {
	out, err := runCommand("git", "-C", gitRoot, "diff", baseBranch, "--", "*.tf", "*.hcl")
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}
	return out, nil
}

// runCommand executes a command and returns its combined output.
func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}

// parseDiff parses unified diff output and extracts resource/module blocks.
func parseDiff(diffOutput string) *DiffResult {
	result := &DiffResult{}
	if diffOutput == "" {
		return result
	}

	lines := strings.Split(diffOutput, "\n")
	var currentFile string
	var currentBlock *blockContext
	seenFiles := map[string]bool{}
	seenResources := map[string]*DiffResource{}
	seenModules := map[string]*DiffModule{}

	for _, line := range lines {
		// Detect file header
		if strings.HasPrefix(line, "diff --git") {
			currentBlock = nil
			continue
		}
		if strings.HasPrefix(line, "+++ b/") {
			currentFile = strings.TrimPrefix(line, "+++ b/")
			if !seenFiles[currentFile] {
				seenFiles[currentFile] = true
				result.FilesChanged = append(result.FilesChanged, currentFile)
			}
			currentBlock = nil
			continue
		}
		if strings.HasPrefix(line, "--- ") || strings.HasPrefix(line, "@@") {
			continue
		}

		// Determine line type
		lineType := "context"
		content := line
		if strings.HasPrefix(line, "+") {
			lineType = "added"
			content = line[1:]
		} else if strings.HasPrefix(line, "-") {
			lineType = "removed"
			content = line[1:]
		} else if len(line) > 0 {
			content = line
			if content[0] == ' ' {
				content = content[1:]
			}
		}

		// Try to detect block headers (resource, module, data)
		if m := resourceRe.FindStringSubmatch(content); m != nil {
			key := m[1] + "." + m[2] + "@" + currentFile
			if _, exists := seenResources[key]; !exists {
				action := "modified"
				if lineType == "added" {
					action = "added"
				} else if lineType == "removed" {
					action = "removed"
				}
				r := &DiffResource{Type: m[1], Name: m[2], FilePath: currentFile, Action: action}
				seenResources[key] = r
				result.Resources = append(result.Resources, *r)
			}
			currentBlock = &blockContext{blockType: "resource", key: key}
			continue
		}

		if m := dataRe.FindStringSubmatch(content); m != nil {
			key := "data." + m[1] + "." + m[2] + "@" + currentFile
			if _, exists := seenResources[key]; !exists {
				action := "modified"
				if lineType == "added" {
					action = "added"
				} else if lineType == "removed" {
					action = "removed"
				}
				r := &DiffResource{Type: "data." + m[1], Name: m[2], FilePath: currentFile, Action: action}
				seenResources[key] = r
				result.Resources = append(result.Resources, *r)
			}
			currentBlock = &blockContext{blockType: "data", key: key}
			continue
		}

		if m := moduleRe.FindStringSubmatch(content); m != nil {
			key := m[1] + "@" + currentFile
			if _, exists := seenModules[key]; !exists {
				action := "modified"
				if lineType == "added" {
					action = "added"
				} else if lineType == "removed" {
					action = "removed"
				}
				mod := &DiffModule{Name: m[1], FilePath: currentFile, Action: action}
				seenModules[key] = mod
				result.Modules = append(result.Modules, *mod)
			}
			currentBlock = &blockContext{blockType: "module", key: key}
			continue
		}

		// Context-aware: if we're inside a known block and see a changed line,
		// the block is being modified
		if currentBlock != nil && (lineType == "added" || lineType == "removed") {
			if currentBlock.blockType == "resource" || currentBlock.blockType == "data" {
				if r, exists := seenResources[currentBlock.key]; exists {
					if r.Action == "modified" || r.Action == "" {
						// Already tracked as modified from context
					}
				}
			}
		}
	}

	// Second pass: detect resources that appear only in context lines but have
	// modifications within their block. We need to check for resource declarations
	// in context lines followed by added/removed lines.
	result.Resources = deduplicateResources(result.Resources)
	result.Modules = deduplicateModules(result.Modules)

	return result
}

type blockContext struct {
	blockType string // "resource", "module", "data"
	key       string
}

func deduplicateResources(resources []DiffResource) []DiffResource {
	seen := map[string]int{}
	var result []DiffResource
	for _, r := range resources {
		key := r.Type + "." + r.Name + "@" + r.FilePath
		if idx, exists := seen[key]; exists {
			// Keep the more specific action
			if result[idx].Action == "modified" && r.Action != "modified" {
				result[idx].Action = r.Action
			}
		} else {
			seen[key] = len(result)
			result = append(result, r)
		}
	}
	return result
}

func deduplicateModules(modules []DiffModule) []DiffModule {
	seen := map[string]int{}
	var result []DiffModule
	for _, m := range modules {
		key := m.Name + "@" + m.FilePath
		if idx, exists := seen[key]; exists {
			if result[idx].Action == "modified" && m.Action != "modified" {
				result[idx].Action = m.Action
			}
		} else {
			seen[key] = len(result)
			result = append(result, m)
		}
	}
	return result
}

// stripAddressIndices removes index brackets from a terraform address.
func stripAddressIndices(address string) string {
	return indexRe.ReplaceAllString(address, "")
}

// extractResourceIdentifier parses a terraform plan address to get type and name.
func extractResourceIdentifier(address string) (string, string) {
	// Strip module prefix(es): module.vpc.module.sub.aws_subnet.public -> aws_subnet.public
	addr := stripAddressIndices(address)
	parts := strings.Split(addr, ".")

	// Walk past module.X pairs to find the resource type.name
	i := 0
	for i < len(parts)-1 {
		if parts[i] == "module" {
			i += 2 // skip "module" and the module name
		} else {
			break
		}
	}

	remaining := parts[i:]
	if len(remaining) >= 2 {
		return remaining[0], remaining[1]
	}
	if len(remaining) == 1 {
		return remaining[0], ""
	}
	return "", ""
}

// extractModuleName returns the first module name from a terraform address.
func extractModuleName(address string) string {
	addr := stripAddressIndices(address)
	parts := strings.Split(addr, ".")
	for i, p := range parts {
		if p == "module" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// ResourceIdentifiers returns a list of "type.name" strings from the diff resources.
func (d *DiffResult) ResourceIdentifiers() []string {
	var ids []string
	for _, r := range d.Resources {
		ids = append(ids, r.Type+"."+r.Name)
	}
	return ids
}

// ModuleNames returns a list of module names from the diff.
func (d *DiffResult) ModuleNames() []string {
	var names []string
	for _, m := range d.Modules {
		names = append(names, m.Name)
	}
	return names
}

// FilePaths returns absolute file paths for the changed files.
func (d *DiffResult) FilePaths(gitRoot string) []string {
	var paths []string
	for _, f := range d.FilesChanged {
		paths = append(paths, filepath.Join(gitRoot, f))
	}
	return paths
}
