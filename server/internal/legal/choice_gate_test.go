package legal_test

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

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// choice_gate_test.go — #794. The enumerator and the engine answer
// "does this prompt stop the table?" with the same function.
//
// They did not. `game.blockingChoiceLocked` lets play continue through
// a `pay_unless` (ADR 0018 §6: the Rhystic tax is another player's
// question, asked after the trigger has already resolved), while
// `legal.anyChoiceOpen` returned an empty move list for every seat
// while any prompt at all was open. So a bot sat on its hands until a
// human answered a tax — the opposite of the ADR's intent, and of what
// a human at the same table may do.
//
// Both now read game.ChoiceBlocksTable, and the last test in this file
// is the mechanism that keeps a NEW kind from re-opening either gap:
// it fails until the kind is classified by the gate and enumerated
// here.

// queueTax puts a Rhystic-shaped tax prompt on `taxed`.
func queueTax(t *testing.T, g *game.Game, taxed uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.QueuePayUnlessForEffect(taxed, uuid.New(), "{1}",
			"Rhystic Study — pay {1}?", nil); err != nil {
			t.Fatalf("QueuePayUnlessForEffect: %v", err)
		}
	})
}

// TestPayUnlessLeavesTheOtherSeatsPlaying is the bug, stated: the
// engine plays on through an unanswered tax, so the enumerator must
// too.
func TestPayUnlessLeavesTheOtherSeatsPlaying(t *testing.T) {
	g := newTable(t)
	me, taxed := g.Seats[0], g.Seats[1]
	// Seat 0 is the active player parked on its draw step, so its
	// ordinary window is a pass plus whatever it holds.
	before := legal.EnumerateFor(g, me.ID)
	if len(before) == 0 {
		t.Fatal("setup: seat 0 had no moves before the tax was queued")
	}

	queueTax(t, g, taxed.ID)

	after := legal.EnumerateFor(g, me.ID)
	if len(after) != len(before) {
		t.Errorf("seat 0 was offered %d moves with a tax open and %d without: %v",
			len(after), len(before), labels(after))
	}
	if !hasLabel(after, "Pass priority") {
		t.Errorf("seat 0 cannot even pass while somebody else owes a tax: %v", labels(after))
	}
	dispatchAll(t, g, me.ID, after)
}

// TestPayUnlessIsStillEnumeratedForTheSeatThatOwesIt — the other half
// of the rule. A non-blocking prompt is still a prompt, and the seat
// holding it must be able to pay or decline; #794's fix must not turn
// "the table plays on" into "nobody is asked".
func TestPayUnlessIsStillEnumeratedForTheSeatThatOwesIt(t *testing.T) {
	g := newTable(t)
	taxed := g.Seats[1]
	queueTax(t, g, taxed.ID)

	moves := legal.EnumerateFor(g, taxed.ID)
	if len(moves) == 0 {
		t.Fatal("the taxed seat was offered nothing at all")
	}
	for _, m := range moves {
		if m.Type != legal.TypeResolveChoice {
			t.Errorf("the taxed seat was offered %q as well as the tax answers", m.Label)
		}
	}
	if !hasLabel(moves, "Rhystic Study — pay {1}?: decline") {
		t.Errorf("no decline answer: %v", labels(moves))
	}
	dispatchAll(t, g, taxed.ID, moves)
}

// TestSacrificePromptStillEmptiesEveryOtherSeat — the gate that #794
// must not widen. A blocking prompt offers its owner its answers and
// every other seat nothing, exactly as before.
func TestSacrificePromptStillEmptiesEveryOtherSeat(t *testing.T) {
	g := newTable(t)
	me, them := g.Seats[0], g.Seats[1]
	battlefieldCard(g, me, creature("Doomed Traveler", "{W}", 1, 1))
	g.WithWriteLock(func() {
		if n := g.PlayerSacrificesForEffect(uuid.New(), me.ID, nil,
			"Fleshbag Marauder — sacrifice a creature"); n != 1 {
			t.Fatalf("queued %d sacrifice prompts, want 1", n)
		}
	})

	mine := legal.EnumerateFor(g, me.ID)
	if len(mine) == 0 {
		t.Fatal("the seat owing the sacrifice was offered nothing")
	}
	for _, m := range mine {
		if m.Type != legal.TypeResolveChoice {
			t.Errorf("the seat owing the sacrifice was offered %q as well", m.Label)
		}
	}
	if moves := legal.EnumerateFor(g, them.ID); len(moves) != 0 {
		t.Errorf("another seat was offered %v while a sacrifice prompt was open", labels(moves))
	}
	dispatchAll(t, g, me.ID, mine)
}

// TestEveryChoiceKindIsClassifiedAndEnumerated is the mechanism.
//
// A new PendingChoiceKind has two obligations and no compiler to
// enforce either: a row in the engine's gate (or the table silently
// takes the deny-by-default answer nobody chose) and a case in
// choiceMoves (or every seat owing one gets an EMPTY move list — the
// #499 / #618 wedge, which is how Door of Destinies and Cavern of
// Souls stopped real tables). Both failures are silent until a game
// stops, so this test reads the kinds straight out of the source and
// fails the moment one is declared without either.
func TestEveryChoiceKindIsClassifiedAndEnumerated(t *testing.T) {
	declared := declaredChoiceKinds(t)
	if len(declared) < 20 {
		t.Fatalf("found only %d PendingChoiceKind constants in ../game — the scanner has stopped working", len(declared))
	}

	classified := map[game.PendingChoiceKind]bool{}
	for _, kind := range game.ClassifiedChoiceKinds() {
		classified[kind] = true
	}
	enumerated := enumeratedChoiceKinds(t)

	names := make([]string, 0, len(declared))
	for name := range declared {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if !classified[declared[name]] {
			t.Errorf("game.%s (%q) has no row in choiceGateDecisions — decide whether it stops the table and say so in server/internal/game/choice_gate.go (ADR 0018 §6)",
				name, declared[name])
		}
		if !enumerated[name] {
			t.Errorf("game.%s (%q) has no case in choiceMoves — a seat owing one would be offered no answer and no pass (#499 / #618); add it to server/internal/legal/choices.go",
				name, declared[name])
		}
	}

	// The reverse direction catches a row or a case left behind by a
	// kind that has been deleted.
	for kind := range classified {
		found := false
		for _, v := range declared {
			if v == kind {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("choiceGateDecisions classifies %q, which no longer exists", kind)
		}
	}
}

// declaredChoiceKinds parses ../game for every `const X PendingChoiceKind
// = "..."` and returns constant name → value. Source scanning rather
// than reflection because Go has no way to enumerate the constants of
// a named string type at runtime, and a hand-kept list here would be
// the third copy of the thing this file exists to stop duplicating.
func declaredChoiceKinds(t *testing.T) map[string]game.PendingChoiceKind {
	t.Helper()
	entries, err := os.ReadDir("../game")
	if err != nil {
		t.Fatalf("read ../game: %v", err)
	}
	fset := gotoken.NewFileSet()
	out := map[string]game.PendingChoiceKind{}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join("../game", name), nil, 0)
		if err != nil {
			t.Fatalf("parse ../game/%s: %v", name, err)
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
				if !ok || ident.Name != "PendingChoiceKind" {
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
					out[constName.Name] = game.PendingChoiceKind(value)
				}
			}
		}
	}
	return out
}

// enumeratedChoiceKinds parses choices.go and returns the set of
// `game.PendingChoiceX` names that appear as a case of choiceMoves'
// `switch c.Kind`. That switch IS the enumerator's coverage, so
// reading it is reading the answer rather than a record of it.
func enumeratedChoiceKinds(t *testing.T) map[string]bool {
	t.Helper()
	fset := gotoken.NewFileSet()
	file, err := parser.ParseFile(fset, "choices.go", nil, 0)
	if err != nil {
		t.Fatalf("parse choices.go: %v", err)
	}
	out := map[string]bool{}
	ast.Inspect(file, func(n ast.Node) bool {
		sw, ok := n.(*ast.SwitchStmt)
		if !ok || !isChoiceKindTag(sw.Tag) {
			return true
		}
		for _, stmt := range sw.Body.List {
			clause, ok := stmt.(*ast.CaseClause)
			if !ok {
				continue
			}
			for _, expr := range clause.List {
				sel, ok := expr.(*ast.SelectorExpr)
				if !ok {
					continue
				}
				if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "game" {
					out[sel.Sel.Name] = true
				}
			}
		}
		return true
	})
	if len(out) == 0 {
		t.Fatal("found no `switch c.Kind` cases in choices.go — the scanner has stopped working")
	}
	return out
}

// isChoiceKindTag matches the `c.Kind` a choice switch keys on.
func isChoiceKindTag(tag ast.Expr) bool {
	sel, ok := tag.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Kind" {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	return ok && ident.Name == "c"
}
