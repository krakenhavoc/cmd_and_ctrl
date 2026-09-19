package game

import (
	"go/ast"
	"go/parser"
	gotoken "go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// replacement_kind_gate_test.go — #982. A ReplacementEventKind has
// obligations at five switches and no compiler to enforce any of them.
//
// The bug that produced this file: `affectedPlayerForEvent` named the
// affected player for draw, life, counter, damage, move and step, and
// had no case for either kind ADR 0061 added. A discard and a token
// creation fell through to "the first gathered effect's controller" —
// whoever happened to sit earliest in battlefield order — so a
// symmetrical replacement (Primal Vigor, madness) would have put the
// CR 616.1 ordering prompt to the wrong player. It was invisible
// because every catalog replacement of those two kinds is
// controller-scoped, which makes the fallback accidentally right.
//
// Every switch below has that property: a missing arm is silently
// WRONG rather than a build error or a crash, and stays invisible until
// the first card whose scope differs from the default. So the test
// reads the kinds and the switches straight out of the source — the
// mechanism `TestEveryChoiceKindIsClassifiedAndEnumerated`
// (internal/legal/choice_gate_test.go) uses for PendingChoiceKind — and
// fails the moment a kind is declared without an arm in each.
//
// It found one on its first run: `abandonZoneRouteLocked` had no arm
// for RepEventKeywordAction, so a scry whose CR 616 prompt was taken
// away dropped its "then draw a card" with the frame.
//
// Two of the five arrived as if-chains rather than switches
// (`finishSettledReplacementLocked`, `abandonZoneRouteLocked`) and were
// rewritten into switches for this, with no change to what they do:
// every condition on either chain was exclusive on Kind, so the arms
// are the chain's branches in the chain's order. "Nothing is owed for
// this kind" is now a written arm rather than the end of a chain.

// replacementKindSwitches is the list. Each entry names a function
// whose `switch ev.Kind` (or `switch kind`) must have an arm for EVERY
// declared ReplacementEventKind, and says what a missing arm costs.
//
// Adding a sixth is one row. A switch that does NOT belong here is one
// that is deliberately partial — `gatherSelfReplacementsLocked`'s
// `ev.Kind != RepEventMove` guard, a card's `AppliesTo` — and the test
// never looks at those, on purpose.
var replacementKindSwitches = []struct {
	file string
	fn   string
	cost string
}{
	{
		file: "replacements.go",
		fn:   "eventKindMatches",
		cost: "no EventKind maps to the kind, so `Watches` never matches it and NO replacement of that kind can ever fire",
	},
	{
		file: "pending_choice.go",
		fn:   "affectedPlayerForEvent",
		cost: "the CR 616.1 ordering prompt goes to the first gathered effect's controller instead of the affected player (#982)",
	},
	{
		file: "pending_choice.go",
		fn:   "applyResolvedReplacementEventLocked",
		cost: "an event of that kind that PAUSED on a prompt resumes into nothing — the mutation never happens",
	},
	{
		file: "pending_choice.go",
		fn:   "finishSettledReplacementLocked",
		cost: "a CANCELLED event of that kind never tells its caller, so a continuation sequenced behind it waits forever (#762, #793, #807, #853)",
	},
	{
		file: "zone_route.go",
		fn:   "abandonZoneRouteLocked",
		cost: "an event of that kind whose prompt is DROPPED or PRUNED never tells its caller, with the same stall (#865)",
	},
}

// TestEveryReplacementEventKindIsSwitchedOn is the mechanism.
func TestEveryReplacementEventKindIsSwitchedOn(t *testing.T) {
	declared := declaredReplacementKinds(t)
	if len(declared) < 8 {
		t.Fatalf("found only %d ReplacementEventKind constants — the scanner has stopped working", len(declared))
	}

	names := make([]string, 0, len(declared))
	for name := range declared {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, sw := range replacementKindSwitches {
		arms := replacementKindArms(t, sw.file, sw.fn)
		if len(arms) == 0 {
			t.Fatalf("%s/%s: found no `switch ev.Kind` arms — the scanner has stopped working", sw.file, sw.fn)
		}
		for _, name := range names {
			if arms[name] {
				continue
			}
			t.Errorf("%s (%q) has no arm in %s (%s): %s",
				name, declared[name], sw.fn, sw.file, sw.cost)
		}
		// The reverse direction catches an arm left behind by a kind
		// that has been deleted. A stale arm is dead code rather than a
		// wrong answer, but it is also the only evidence that the
		// deletion was incomplete.
		armNames := make([]string, 0, len(arms))
		for name := range arms {
			armNames = append(armNames, name)
		}
		sort.Strings(armNames)
		for _, name := range armNames {
			if _, ok := declared[name]; !ok {
				t.Errorf("%s (%s) has an arm for %s, which is not a declared ReplacementEventKind",
					sw.fn, sw.file, name)
			}
		}
	}
}

// declaredReplacementKinds parses this package for every
// `const X ReplacementEventKind = "..."` and returns constant name →
// value. Source scanning rather than reflection because Go cannot
// enumerate the constants of a named string type at runtime, and a
// hand-kept list here would be the sixth copy of the thing this file
// exists to stop drifting.
func declaredReplacementKinds(t *testing.T) map[string]ReplacementEventKind {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read .: %v", err)
	}
	fset := gotoken.NewFileSet()
	out := map[string]ReplacementEventKind{}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != gotoken.CONST {
				continue
			}
			for _, spec := range gen.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				ident, ok := vs.Type.(*ast.Ident)
				if !ok || ident.Name != "ReplacementEventKind" {
					continue
				}
				for i, constName := range vs.Names {
					if i >= len(vs.Values) {
						continue
					}
					lit, ok := vs.Values[i].(*ast.BasicLit)
					if !ok || lit.Kind != gotoken.STRING {
						continue
					}
					value, err := strconv.Unquote(lit.Value)
					if err != nil {
						t.Fatalf("unquote %s = %s: %v", constName.Name, lit.Value, err)
					}
					out[constName.Name] = ReplacementEventKind(value)
				}
			}
		}
	}
	return out
}

// replacementKindArms parses one function and returns the set of
// `RepEventXxx` names that appear as a case of a switch on a
// ReplacementEventKind. The switch IS the coverage, so reading it is
// reading the answer rather than a record of it.
func replacementKindArms(t *testing.T, file, fn string) map[string]bool {
	t.Helper()
	fset := gotoken.NewFileSet()
	parsed, err := parser.ParseFile(fset, filepath.Clean(file), nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", file, err)
	}
	out := map[string]bool{}
	for _, decl := range parsed.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Name == nil || fd.Name.Name != fn || fd.Body == nil {
			continue
		}
		ast.Inspect(fd.Body, func(n ast.Node) bool {
			sw, ok := n.(*ast.SwitchStmt)
			if !ok || !isReplacementKindTag(sw.Tag) {
				return true
			}
			for _, stmt := range sw.Body.List {
				clause, ok := stmt.(*ast.CaseClause)
				if !ok {
					continue
				}
				for _, expr := range clause.List {
					if ident, ok := expr.(*ast.Ident); ok && strings.HasPrefix(ident.Name, "RepEvent") {
						out[ident.Name] = true
					}
				}
			}
			return true
		})
		return out
	}
	t.Fatalf("no func %s in %s", fn, file)
	return nil
}

// isReplacementKindTag matches the two spellings the five switches use
// for their subject: `ev.Kind` on a *ReplacementEvent, and the bare
// `kind` parameter eventKindMatches takes.
func isReplacementKindTag(tag ast.Expr) bool {
	switch t := tag.(type) {
	case *ast.SelectorExpr:
		if t.Sel == nil || t.Sel.Name != "Kind" {
			return false
		}
		ident, ok := t.X.(*ast.Ident)
		return ok && (ident.Name == "ev" || ident.Name == "out")
	case *ast.Ident:
		return t.Name == "kind"
	}
	return false
}
