package effects

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// gated_event_kinds_guard_test.go — #4404, case 2. A CR 616 frame over a
// mana production, a step transition, an extra turn, a counter-spell or a
// dice roll is never written (replacementEventFields gates their fields),
// and docs/snapshots.md closes the case because none of those events can
// pause with the catalog as it is:
//
//   - a mana production, an extra turn's start and a dice roll are opened
//     with mustSettleNow (CR 605.3b for the mana: a mana ability resolves
//     immediately), so their window applies the gathered order and never
//     asks;
//   - every step-transition replacement is a PureCancel, and two pure
//     cancels are one answer (allPureCancels), so they never ask either;
//   - Guile's is the only counter-spell replacement, and it applies only
//     to its own controller's counters, so two never meet on one event
//     except under one controller, where they are one modification (#792).
//
// This reads every ReplacementEffect and ReplacementEvent literal in the
// catalog and the engine and fails when any of that stops being true. If
// it fails, the kind can now pause: key its fields in
// replacement_resume_keyed.go (a frame of its own, as #4362 keyed combat
// damage) before relaxing this test, or the new prompt is not a restore
// point.

func TestGatedReplacementEventKindsCannotPause(t *testing.T) {
	var problems []string
	counterSpell := 0
	settled := map[string]int{}
	for _, dir := range []string{".", filepath.Join("..", "..", "game")} {
		fset := token.NewFileSet()
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			name := e.Name()
			if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			checkEffect := func(lit *ast.CompositeLit) {
				keys := literalKeys(lit)
				where := fset.Position(lit.Pos()).String()
				if mentions(keys["Watches"], "EventStepTransition") && !isTrue(keys["PureCancel"]) {
					problems = append(problems, where+": a step-transition replacement that is not a PureCancel can pause the step on a CR 616 prompt, which is not a restore point")
				}
				if mentions(keys["Watches"], "EventCounterSpell") {
					counterSpell++
					if name != "counter_replacements.go" {
						problems = append(problems, where+": a second counter-spell replacement can meet Guile's on one event and pause it on a CR 616 prompt, which is not a restore point")
					}
				}
			}
			ast.Inspect(f, func(n ast.Node) bool {
				lit, ok := n.(*ast.CompositeLit)
				if !ok {
					return true
				}
				switch {
				case isReplacementEffectType(lit.Type):
					checkEffect(lit)
				case isReplacementEffectSlice(lit.Type):
					for _, el := range lit.Elts {
						if inner, ok := el.(*ast.CompositeLit); ok && inner.Type == nil {
							checkEffect(inner)
						}
					}
				case isReplacementEventType(lit.Type):
					keys := literalKeys(lit)
					for _, kind := range []string{"RepEventProduceMana", "RepEventExtraTurn", "RepEventRollDice"} {
						if !mentions(keys["Kind"], kind) {
							continue
						}
						settled[kind]++
						if !isTrue(keys["mustSettleNow"]) {
							problems = append(problems, fset.Position(lit.Pos()).String()+": a "+kind+" event opened without mustSettleNow can pause on a CR 616 prompt, which is not a restore point")
						}
					}
				}
				return true
			})
		}
	}
	if counterSpell != 1 {
		t.Errorf("%d counter-spell replacement literals, want Guile's one: the guard has lost sight of it", counterSpell)
	}
	for _, kind := range []string{"RepEventProduceMana", "RepEventExtraTurn", "RepEventRollDice"} {
		if settled[kind] == 0 {
			t.Errorf("no %s event literal found: the guard has lost sight of where the engine opens one", kind)
		}
	}
	sort.Strings(problems)
	for _, p := range problems {
		t.Error(p)
	}
}

// literalKeys is a keyed composite literal's fields by name.
func literalKeys(lit *ast.CompositeLit) map[string]ast.Expr {
	keys := map[string]ast.Expr{}
	for _, el := range lit.Elts {
		if kv, ok := el.(*ast.KeyValueExpr); ok {
			if id, ok := kv.Key.(*ast.Ident); ok {
				keys[id.Name] = kv.Value
			}
		}
	}
	return keys
}

func isTrue(e ast.Expr) bool {
	id, ok := e.(*ast.Ident)
	return ok && id.Name == "true"
}

func isReplacementEventType(e ast.Expr) bool {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name == "ReplacementEvent"
	case *ast.SelectorExpr:
		return x.Sel.Name == "ReplacementEvent"
	}
	return false
}
