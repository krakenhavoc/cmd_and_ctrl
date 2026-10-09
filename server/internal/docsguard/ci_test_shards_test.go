package docsguard_test

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

// shardScript is scripts/ci-test-shard.sh relative to this package.
const shardScript = "../../../scripts/ci-test-shard.sh"

// ciWorkflow is the CI workflow relative to this package.
const ciWorkflow = "../../../.github/workflows/ci-cd.yml"

// matrixLine is the one line in ci-cd.yml that lists server-test's shards.
var matrixLine = regexp.MustCompile(`(?m)^\s*shard:\s*\[([^\]]*)\]\s*$`)

// TestCITestShardsCoverEveryPackageOnce is #2766's guard on the sharded
// `server-test` job: no package is lost and none runs twice. The shards
// come from scripts/ci-test-shard.sh, which pins the heavy packages and
// puts everything else in `rest`; this test checks the union of its
// shards against `go list ./...` and that the workflow's matrix runs
// exactly the shards the script defines, in its order.
func TestCITestShardsCoverEveryPackageOnce(t *testing.T) {
	names := lines(t, runShardScript(t, "--names"))
	if len(names) < 2 {
		t.Fatalf("%s --names listed %v; expected at least one pinned shard and rest", shardScript, names)
	}

	listCmd := exec.Command("go", "list", "./...")
	listCmd.Dir = serverRoot
	out, err := listCmd.Output()
	if err != nil {
		t.Fatalf("go list ./...: %v", err)
	}
	want := lines(t, string(out))

	seen := map[string]string{}
	for _, name := range names {
		pkgs := lines(t, runShardScript(t, name))
		if len(pkgs) == 0 {
			t.Errorf("shard %q runs no packages", name)
		}
		for _, pkg := range pkgs {
			if other, dup := seen[pkg]; dup {
				t.Errorf("%s is in both shard %q and shard %q", pkg, other, name)
			}
			seen[pkg] = name
		}
	}
	for _, pkg := range want {
		if _, ok := seen[pkg]; !ok {
			t.Errorf("%s is in go list ./... but in no CI test shard", pkg)
		}
	}
	listed := map[string]bool{}
	for _, pkg := range want {
		listed[pkg] = true
	}
	for pkg, name := range seen {
		if !listed[pkg] {
			t.Errorf("shard %q runs %s, which go list ./... does not list", name, pkg)
		}
	}

	raw, err := os.ReadFile(ciWorkflow)
	if err != nil {
		t.Fatalf("read %s: %v", ciWorkflow, err)
	}
	m := matrixLine.FindAllStringSubmatch(string(raw), -1)
	if len(m) != 1 {
		t.Fatalf("%s should have exactly one `shard: [...]` matrix line (server-test's), found %d", ciWorkflow, len(m))
	}
	var matrix []string
	for _, s := range strings.Split(m[0][1], ",") {
		matrix = append(matrix, strings.Trim(strings.TrimSpace(s), `"'`))
	}
	if strings.Join(matrix, ",") != strings.Join(names, ",") {
		t.Errorf("server-test's matrix in %s is %v, but %s --names is %v; a shard the matrix "+
			"does not run is a set of packages CI never tests", ciWorkflow, matrix, shardScript, names)
	}
}

func runShardScript(t *testing.T, arg string) string {
	t.Helper()
	cmd := exec.Command("bash", shardScript, arg)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("%s %s: %v\n%s", shardScript, arg, err, stderr.String())
	}
	return string(out)
}

func lines(t *testing.T, s string) []string {
	t.Helper()
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out
}
