package effects

import (
	"go/ast"
	"go/token"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// discover_guard_test.go — Spec.Discovers and the source must agree
// (ADR 0099). A card file that builds a Discover{…} or calls
// DiscoverN(…) declares Discovers: true on its Spec, and a Spec that
// declares it has one of the two somewhere in its file. The
// declaration is what cards/coverage reads to catch a stale "discover
// isn't implemented" caveat, so a card that discovers without saying
// so would hide exactly the drift the probe exists for.
//
// Per FILE rather than per Spec literal, because a catalog card is one
// file (AGENTS.md §7) and its helpers live beside its Register call. A
// file with several Specs (a double-faced card's two faces, a cycle
// written as a table) must declare it on the Spec whose text discovers;
// the check is that at least one does.

func fileCallsDiscover(file *ast.File) bool {
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CompositeLit:
			if id, ok := x.Type.(*ast.Ident); ok && id.Name == "Discover" {
				found = true
			}
		case *ast.CallExpr:
			if id, ok := x.Fun.(*ast.Ident); ok && id.Name == "DiscoverN" {
				found = true
			}
		}
		return !found
	})
	return found
}

func specDeclaresDiscovers(lit *ast.CompositeLit) bool {
	for _, el := range lit.Elts {
		kv, ok := el.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != "Discovers" {
			continue
		}
		val, ok := kv.Value.(*ast.Ident)
		return ok && val.Name == "true"
	}
	return false
}

func TestSpecDiscoversMatchesTheSource(t *testing.T) {
	fset := token.NewFileSet()
	files := parseCatalogSources(t, fset)
	if len(files) == 0 {
		t.Fatal("scanned no catalog files — a lint that reads nothing passes everything")
	}
	var findings []string
	for name, file := range files {
		specs := registeredSpecs(file)
		if len(specs) == 0 {
			continue
		}
		declared := false
		line := 0
		for _, lit := range specs {
			if specDeclaresDiscovers(lit) {
				declared = true
				line = fset.Position(lit.Pos()).Line
			}
		}
		calls := fileCallsDiscover(file)
		switch {
		case calls && !declared:
			findings = append(findings, name+" discovers but no Spec in it declares Discovers: true")
		case declared && !calls:
			findings = append(findings, name+":"+strconv.Itoa(line)+" declares Discovers: true but never builds a Discover or calls DiscoverN")
		}
	}
	if len(findings) == 0 {
		return
	}
	sort.Strings(findings)
	t.Errorf("Spec.Discovers and the source disagree (ADR 0099):\n  %s", strings.Join(findings, "\n  "))
}
