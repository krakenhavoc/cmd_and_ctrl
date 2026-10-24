package aiseat_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// runner_lifecycle_guard_test.go keeps #4219 fixed.
//
// #4219 was a data race between a bot runner that an earlier test had
// cancelled but not waited for, still inside a step and reading
// game.CatalogTriggers, and a later test that swapped that hook
// (installTriggerLoop). Any test that swaps a process-wide catalog
// hook is exposed to it, and the only protection that does not depend
// on every author remembering is that no test starts a runner it does
// not also wait for.
//
// The wait lives in three helpers: startBot, newManager and newTierManager
// in runner_test.go (for the aiseat_test package) and startRunner in
// pass_pacing_internal_test.go (for package aiseat). Each registers a
// t.Cleanup that cancels the runner and waits for Done, which runs
// after the test's defers and before the next test begins. This test
// fails when a test file starts a runner or a manager any other way.

// sanctionedStarters are the helper functions allowed to make the raw
// call, because they are the ones that clean up after it.
var sanctionedStarters = map[string]bool{
	"startBot":       true,
	"newManager":     true,
	"newTierManager": true,
	"startRunner":    true, // startBot for the internal tests
}

// rawRunnerStarts returns "func: call at file:line" for every call in
// src that starts a runner or a manager outside a sanctioned helper:
// the qualified aiseat.Start, aiseat.NewManager and
// aiseat.NewManagerWithConfig (external tests), and the bare Start
// (internal tests). A nil src reads the file.
func rawRunnerStarts(t *testing.T, name string, src any) []string {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, name, src, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	var found []string
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil || sanctionedStarters[fn.Name.Name] {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			var callee string
			switch fun := call.Fun.(type) {
			case *ast.SelectorExpr:
				pkg, ok := fun.X.(*ast.Ident)
				if !ok || pkg.Name != "aiseat" {
					return true
				}
				switch fun.Sel.Name {
				case "Start", "NewManager", "NewManagerWithConfig":
					callee = "aiseat." + fun.Sel.Name
				}
			case *ast.Ident:
				// Only an internal test can call Start unqualified.
				if fun.Name == "Start" && f.Name.Name == "aiseat" {
					callee = "Start"
				}
			}
			if callee != "" {
				found = append(found, fn.Name.Name+": "+callee+" at "+name+":"+strconv.Itoa(fset.Position(call.Pos()).Line))
			}
			return true
		})
	}
	return found
}

func TestNoTestStartsARunnerItDoesNotStop(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read the package directory: %v", err)
	}
	var bad []string
	files := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		files++
		bad = append(bad, rawRunnerStarts(t, e.Name(), nil)...)
	}
	if files < 20 {
		t.Fatalf("scanned %d test files; the guard is not looking at the package", files)
	}
	sort.Strings(bad)
	for _, b := range bad {
		t.Errorf("%s: use startBot, newManager, newTierManager or startRunner, which stop the runner when the test ends (#4219)", b)
	}
}

// The guard has to be able to fail: a fixture with a raw start in each
// form it knows, and one inside a sanctioned helper.
func TestRunnerLifecycleGuardSeesARawStart(t *testing.T) {
	const external = `package aiseat_test
func TestLeaks() {
	r := aiseat.Start(ctx, room, seat, pol, cfg, nil, log)
	m := aiseat.NewManagerWithConfig(nil, cfg, log)
	n := aiseat.NewManager(nil, log)
	g.Start(rng) // a Game, not a runner
}
func startBot() { aiseat.Start(ctx, room, seat, pol, cfg, nil, log) }
`
	const internal = `package aiseat
func TestLeaksToo() { r := Start(ctx, room, seat, pol, cfg, nil, log) }
func startRunner() { Start(ctx, room, seat, pol, cfg, nil, log) }
`
	if got := rawRunnerStarts(t, "external.go", external); len(got) != 3 {
		t.Errorf("external fixture: %d raw starts found, want 3: %v", len(got), got)
	}
	if got := rawRunnerStarts(t, "internal.go", internal); len(got) != 1 {
		t.Errorf("internal fixture: %d raw starts found, want 1: %v", len(got), got)
	}
}
