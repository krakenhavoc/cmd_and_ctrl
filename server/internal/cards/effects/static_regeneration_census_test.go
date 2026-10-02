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

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// static_regeneration_census_test.go — ADR 0108 §2 decision 2 (#1887).
//
// CR 701.19c's "can't be regenerated" is read at one gate
// (game.RegenerationAllowedForEffect for a catalog static, the
// regeneration built-in's own AppliesTo for a shield). A regeneration
// that forgets to ask it regenerates a creature Incinerate said can't
// be, and nothing else in the tree would ever go red.
//
// So this test reads every ReplacementEffect literal in the catalog and
// in the engine, straight out of the source (the mechanism
// prevention_flag_census_test.go uses), and fails one whose Replace
// regenerates and whose AppliesTo does not ask the gate. It also fails
// a catalog file that cancels a destruction by hand: a static
// regeneration is RegenerateIfThisWouldBeDestroyed, or it calls
// RegenerateInsteadForEffect.

// regeneratingCalls are the calls that perform a regeneration.
var regeneratingCalls = map[string]bool{
	"RegenerateInsteadForEffect":    true,
	"applyRegenerationShieldLocked": true,
	"regenerateLocked":              true,
}

// regenerationGates are the calls that ask the gate.
var regenerationGates = map[string]bool{
	"RegenerationAllowedForEffect":    true,
	"regenerationShieldAppliesLocked": true,
	"regenerationRefusedLocked":       true,
}

func TestEveryRegenerationAsksTheGate(t *testing.T) {
	var problems []string
	seen := 0
	for _, dir := range []string{".", filepath.Join("..", "..", "game")} {
		p, n := regenerationCensus(t, dir)
		problems = append(problems, p...)
		seen += n
	}
	if seen < 2 {
		t.Fatalf("the census saw %d regenerating replacements; the built-in shield and RegenerateIfThisWouldBeDestroyed are two — the scan has rotted", seen)
	}
	sort.Strings(problems)
	for _, p := range problems {
		t.Error(p)
	}
}

func regenerationCensus(t *testing.T, dir string) (problems []string, seen int) {
	t.Helper()
	fset := token.NewFileSet()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var files []*ast.File
	funcs := map[string]*ast.FuncDecl{}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, f)
		for _, d := range f.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok {
				funcs[fd.Name.Name] = fd
			}
		}
	}
	check := func(lit *ast.CompositeLit) {
		keys := map[string]ast.Expr{}
		for _, el := range lit.Elts {
			if kv, ok := el.(*ast.KeyValueExpr); ok {
				if id, ok := kv.Key.(*ast.Ident); ok {
					keys[id.Name] = kv.Value
				}
			}
		}
		if !callsAnyOf(keys["Replace"], regeneratingCalls, funcs) {
			return
		}
		seen++
		if !callsAnyOf(keys["AppliesTo"], regenerationGates, funcs) {
			problems = append(problems, fset.Position(lit.Pos()).String()+
				": a replacement that regenerates must ask the regeneration gate in its AppliesTo (CR 701.19c, ADR 0108 §2)")
		}
	}
	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			lit, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}
			switch {
			case isReplacementEffectType(lit.Type):
				check(lit)
			case isReplacementEffectSlice(lit.Type):
				for _, el := range lit.Elts {
					if inner, ok := el.(*ast.CompositeLit); ok && inner.Type == nil {
						check(inner)
					}
				}
			}
			return true
		})
	}
	return problems, seen
}

// callsAnyOf reports whether the expression calls one of `names`,
// directly or through a function or method of the same package it
// calls (followed one level at a time, to a fixed depth).
func callsAnyOf(e ast.Expr, names map[string]bool, funcs map[string]*ast.FuncDecl) bool {
	if e == nil {
		return false
	}
	visited := map[string]bool{}
	var walk func(n ast.Node, depth int) bool
	walk = func(n ast.Node, depth int) bool {
		found := false
		ast.Inspect(n, func(n ast.Node) bool {
			if found {
				return false
			}
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			var name string
			switch fn := call.Fun.(type) {
			case *ast.Ident:
				name = fn.Name
			case *ast.SelectorExpr:
				name = fn.Sel.Name
			}
			if names[name] {
				found = true
				return false
			}
			if fd := funcs[name]; fd != nil && fd.Body != nil && depth < 4 && !visited[name] {
				visited[name] = true
				if walk(fd.Body, depth+1) {
					found = true
					return false
				}
			}
			return true
		})
		return found
	}
	return walk(e, 0)
}

// RegenerateIfThisWouldBeDestroyed, on a test card: a destruction is
// replaced by a regeneration with no shield, every time; a sacrifice is
// not; and "can't be regenerated this turn" stops it (CR 701.19b, c).
func TestStaticRegenerationRegeneratesEveryDestructionUntilItCant(t *testing.T) {
	const oracle = "test-adr0108-static-regeneration"
	registerForTest(t, Spec{
		OracleID:     oracle,
		Name:         "Test Nimbus",
		Replacements: []game.ReplacementEffect{RegenerateIfThisWouldBeDestroyed()},
	})
	g := newCatalogGame(t)
	opp := g.Seats[1].ID
	cleric := pushCatalogPermanent(g, opp, "Test Nimbus", "Creature — Human Cleric", oracle, false)
	for i := 0; i < 2; i++ {
		p1Destroy(t, g, cleric)
		p1WantZone(t, g, cleric, game.ZoneBattlefield, "the regenerating cleric")
	}
	g.WithWriteLock(func() { g.CantBeRegeneratedThisTurnForEffect(uuid.Nil, cleric, "Furnace Brood") })
	p1Destroy(t, g, cleric)
	p1WantZone(t, g, cleric, game.ZoneGraveyard, "the cleric that can't be regenerated")

	other := pushCatalogPermanent(g, opp, "Test Nimbus", "Creature — Human Cleric", oracle, false)
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(other); err != nil {
			t.Fatalf("sacrifice: %v", err)
		}
	})
	p1WantZone(t, g, other, game.ZoneGraveyard, "a sacrificed cleric")
}
