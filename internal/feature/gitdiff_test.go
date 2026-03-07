package feature

import (
	"os"
	"testing"
)

func TestParseDiff_ResourceBlocks(t *testing.T) {
	diffOutput := `diff --git a/main.tf b/main.tf
--- a/main.tf
+++ b/main.tf
@@ -1,5 +1,10 @@
+resource "aws_s3_bucket" "new_bucket" {
+  bucket = "my-new-bucket"
+}
+
 resource "aws_instance" "web" {
-  instance_type = "t2.micro"
+  instance_type = "t2.small"
 }
`
	result := parseDiff(diffOutput)

	if len(result.Resources) < 2 {
		t.Fatalf("expected at least 2 resources, got %d", len(result.Resources))
	}

	// Check added resource
	found := false
	for _, r := range result.Resources {
		if r.Type == "aws_s3_bucket" && r.Name == "new_bucket" {
			found = true
			if r.Action != "added" {
				t.Errorf("expected action 'added', got %s", r.Action)
			}
			if r.FilePath != "main.tf" {
				t.Errorf("expected file 'main.tf', got %s", r.FilePath)
			}
		}
	}
	if !found {
		t.Error("expected to find aws_s3_bucket.new_bucket in diff resources")
	}

	// Check modified resource
	found = false
	for _, r := range result.Resources {
		if r.Type == "aws_instance" && r.Name == "web" {
			found = true
			if r.Action != "modified" {
				t.Errorf("expected action 'modified', got %s", r.Action)
			}
		}
	}
	if !found {
		t.Error("expected to find aws_instance.web in diff resources")
	}
}

func TestParseDiff_ModuleBlocks(t *testing.T) {
	diffOutput := `diff --git a/modules.tf b/modules.tf
--- a/modules.tf
+++ b/modules.tf
@@ -1,3 +1,8 @@
+module "vpc" {
+  source = "./modules/vpc"
+  cidr   = "10.0.0.0/16"
+}
+
 module "app" {
-  version = "1.0"
+  version = "2.0"
 }
`
	result := parseDiff(diffOutput)

	if len(result.Modules) < 2 {
		t.Fatalf("expected at least 2 modules, got %d", len(result.Modules))
	}

	found := false
	for _, m := range result.Modules {
		if m.Name == "vpc" {
			found = true
			if m.Action != "added" {
				t.Errorf("expected action 'added', got %s", m.Action)
			}
		}
	}
	if !found {
		t.Error("expected to find module 'vpc' in diff modules")
	}

	found = false
	for _, m := range result.Modules {
		if m.Name == "app" {
			found = true
			if m.Action != "modified" {
				t.Errorf("expected action 'modified', got %s", m.Action)
			}
		}
	}
	if !found {
		t.Error("expected to find module 'app' in diff modules")
	}
}

func TestParseDiff_MultipleFiles(t *testing.T) {
	diffOutput := `diff --git a/network.tf b/network.tf
--- a/network.tf
+++ b/network.tf
@@ -1,3 +1,3 @@
 resource "aws_vpc" "main" {
-  cidr_block = "10.0.0.0/16"
+  cidr_block = "10.1.0.0/16"
 }
diff --git a/compute.tf b/compute.tf
--- a/compute.tf
+++ b/compute.tf
@@ -1,3 +1,3 @@
 resource "aws_instance" "app" {
-  ami = "ami-old"
+  ami = "ami-new"
 }
`
	result := parseDiff(diffOutput)

	if len(result.FilesChanged) != 2 {
		t.Errorf("expected 2 files changed, got %d", len(result.FilesChanged))
	}
	if len(result.Resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(result.Resources))
	}

	// Check files
	fileSet := map[string]bool{}
	for _, f := range result.FilesChanged {
		fileSet[f] = true
	}
	if !fileSet["network.tf"] || !fileSet["compute.tf"] {
		t.Errorf("expected network.tf and compute.tf in files, got %v", result.FilesChanged)
	}
}

func TestParseDiff_EmptyDiff(t *testing.T) {
	result := parseDiff("")
	if len(result.Resources) != 0 {
		t.Errorf("expected 0 resources, got %d", len(result.Resources))
	}
	if len(result.Modules) != 0 {
		t.Errorf("expected 0 modules, got %d", len(result.Modules))
	}
	if len(result.FilesChanged) != 0 {
		t.Errorf("expected 0 files, got %d", len(result.FilesChanged))
	}
}

func TestParseDiff_ModifiedLinesInsideBlock(t *testing.T) {
	diffOutput := `diff --git a/main.tf b/main.tf
--- a/main.tf
+++ b/main.tf
@@ -1,6 +1,6 @@
 resource "aws_security_group" "web" {
   name = "web-sg"
-  description = "Old description"
+  description = "New description"

   ingress {
     from_port = 80
`
	result := parseDiff(diffOutput)

	if len(result.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(result.Resources))
	}
	if result.Resources[0].Type != "aws_security_group" || result.Resources[0].Name != "web" {
		t.Errorf("expected aws_security_group.web, got %s.%s", result.Resources[0].Type, result.Resources[0].Name)
	}
	if result.Resources[0].Action != "modified" {
		t.Errorf("expected action 'modified', got %s", result.Resources[0].Action)
	}
}

func TestParseDiff_RemovedResource(t *testing.T) {
	diffOutput := `diff --git a/main.tf b/main.tf
--- a/main.tf
+++ b/main.tf
@@ -1,4 +1,0 @@
-resource "aws_s3_bucket" "old" {
-  bucket = "old-bucket"
-  acl    = "private"
-}
`
	result := parseDiff(diffOutput)

	if len(result.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(result.Resources))
	}
	if result.Resources[0].Action != "removed" {
		t.Errorf("expected action 'removed', got %s", result.Resources[0].Action)
	}
}

func TestParseDiff_DataSource(t *testing.T) {
	diffOutput := `diff --git a/data.tf b/data.tf
--- a/data.tf
+++ b/data.tf
@@ -1,3 +1,3 @@
 data "aws_ami" "ubuntu" {
-  most_recent = false
+  most_recent = true
 }
`
	result := parseDiff(diffOutput)

	found := false
	for _, r := range result.Resources {
		if r.Type == "data.aws_ami" && r.Name == "ubuntu" {
			found = true
		}
	}
	if !found {
		t.Error("expected to find data.aws_ami.ubuntu in diff resources")
	}
}

func TestFindGitRoot_NotAGitRepo(t *testing.T) {
	dir := t.TempDir()
	_, err := findGitRoot(dir)
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}

func TestAnalyzeGitDiff_NotAGitRepo(t *testing.T) {
	dir := t.TempDir()
	result := AnalyzeGitDiff(dir, "main")
	if result.Error == "" {
		t.Error("expected error in result for non-git directory")
	}
	if len(result.Resources) != 0 {
		t.Error("expected empty resources for non-git directory")
	}
}

func TestAnalyzeGitDiff_RealGitRepo(t *testing.T) {
	dir := t.TempDir()

	// Initialize git repo
	runGit := func(args ...string) {
		t.Helper()
		cmd := append([]string{"-C", dir}, args...)
		out, err := runCommand("git", cmd...)
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	runGit("init")
	runGit("config", "user.email", "test@test.com")
	runGit("config", "user.name", "Test")

	// Create initial commit on main
	if err := os.WriteFile(dir+"/main.tf", []byte(`resource "aws_instance" "web" {
  instance_type = "t2.micro"
}
`), 0644); err != nil {
		t.Fatal(err)
	}
	runGit("add", ".")
	runGit("commit", "-m", "initial")
	runGit("branch", "-M", "main")

	// Modify the file (uncommitted change)
	if err := os.WriteFile(dir+"/main.tf", []byte(`resource "aws_instance" "web" {
  instance_type = "t2.small"
}

resource "aws_s3_bucket" "new" {
  bucket = "test"
}
`), 0644); err != nil {
		t.Fatal(err)
	}

	result := AnalyzeGitDiff(dir, "main")
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if len(result.FilesChanged) == 0 {
		t.Error("expected files changed")
	}
	if len(result.Resources) == 0 {
		t.Error("expected resources in diff")
	}
}

func TestStripAddressIndices(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"aws_instance.web[0]", "aws_instance.web"},
		{"module.vpc.aws_subnet.public[\"us-east-1a\"]", "module.vpc.aws_subnet.public"},
		{"aws_instance.web", "aws_instance.web"},
		{"module.app[0].aws_instance.web[1]", "module.app.aws_instance.web"},
		{"module.vpc[\"main\"].aws_route_table.private[0]", "module.vpc.aws_route_table.private"},
	}

	for _, tt := range tests {
		got := stripAddressIndices(tt.input)
		if got != tt.expected {
			t.Errorf("stripAddressIndices(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
