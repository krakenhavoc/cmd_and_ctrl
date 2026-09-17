package effects

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// life_continuation_guard_test.go is the lint #793 could not get from
// `go vet`, extended by #807 to the other half of the same mistake.
//
// Every life change runs the CR 614 window (#482) and so does every
// damage event (S22 / S30), which means either can PAUSE on a CR 616
// ordering prompt when two different replacements apply to it. A card
// that changes a player's life — or deals them damage — and then reads
// a total back to find out how much moved ("each opponent loses X life,
// you gain life equal to the life lost this way"; "deals 1 damage to
// each opponent, you gain life equal to the damage dealt this way")
// reads it before the prompt is answered and sees nothing. It compiles,
// it passes every test played on a table with fewer than two matching
// replacements, and it is wrong.
//
// The right shape is the continuation: ChangePlayerLifeThenForEffect /
// LoseLifeEachThenForEffect on the life side, and
// DealDamageToPlayerThenForEffect / DealDamageToCreatureThenForEffect /
// DealDamageEachThenForEffect on the damage side. The engine runs
// `then` with the amount that ACTUALLY moved, whether that is
// immediately or after the prompt.
//
// So this test reads the catalog's own source and fails on the shape
// that was wrong: a total read back after a change in the same
// function. It is a source scan rather than a behaviour test on
// purpose — the behaviour test needs two replacements on the table to
// catch anything, which is exactly the board nobody writes a test for.
//
// ONE SCANNER, TWO PATTERNS. The life pattern and the damage pattern
// differ only in which calls start the clock and which field reads stop
// it, so readBacksIn takes the pattern as data. A third pattern (a
// counter read-back, say) is a table entry, not another file.
//
// FALSE POSITIVES ARE EXPECTED AND CHEAP. `.Life` is also a field on
// several cost structs, and a loop that reads one player's life and
// then changes another's is fine. Add an entry to the pattern's
// allowlist with a reason; that is the interruption this test is for.

// readBackPattern is one "mutate, then read the result back" shape.
type readBackPattern struct {
	// name is what the failure calls this pattern.
	name string

	// calls are the engine entry points that start the clock: they
	// move the thing being read and they may pause. The *Then* forms
	// and the cost forms are deliberately absent — they are the
	// answer, not the problem.
	calls map[string]bool

	// primitives are catalog primitive structs whose `.Apply(ctx)`
	// reaches one of those entry points — `DealDamage{…}.Apply(ctx)`
	// is how most of the catalog deals damage, and it has to start the
	// clock exactly as the direct call does.
	primitives map[string]bool

	// reads are the field names whose read AFTER one of those calls is
	// the bug.
	reads map[string]bool

	// allowlist exempts one "<file>:<line>" read, with the reason it
	// is not the bug. Keep every entry justified: an unexplained entry
	// is how the lint stops meaning anything.
	allowlist map[string]string

	// advice is printed with a finding: what to use instead.
	advice string
}

// lifeReadBacks is #793's pattern.
var lifeReadBacks = readBackPattern{
	name: "life",
	calls: map[string]bool{
		"ChangePlayerLifeForEffect": true,
		"ChangePlayerLife":          true,
	},
	primitives: map[string]bool{"GainLife": true},
	reads:      map[string]bool{"Life": true},
	allowlist:  map[string]string{},
	advice: "Use ChangePlayerLifeThenForEffect (one player) or LoseLifeEachThenForEffect (several) " +
		"and take the amount from the continuation",
}

// damageReadBacks is #807's pattern: the same mistake with
// DealDamageToPlayerForEffect in place of ChangePlayerLifeForEffect.
// `.Life` is on the list because that is how Creeping Bloodsucker read
// damage back — damage to a player is a life change you can watch —
// and `.DamageMarked` because that is how a card would read damage back
// off a creature.
var damageReadBacks = readBackPattern{
	name: "damage",
	calls: map[string]bool{
		"DealDamageToPlayerForEffect":   true,
		"DealDamageToCreatureForEffect": true,
	},
	primitives: map[string]bool{"DealDamage": true},
	reads:      map[string]bool{"Life": true, "DamageMarked": true},
	allowlist:  map[string]string{},
	advice: "Use DealDamageToPlayerThenForEffect / DealDamageToCreatureThenForEffect (one target) " +
		"or DealDamageEachThenForEffect (several) and take the amount from the continuation",
}

func TestNoReadBackWithoutAContinuation(t *testing.T) {
	for _, pat := range []readBackPattern{lifeReadBacks, damageReadBacks} {
		t.Run(pat.name, func(t *testing.T) {
			findings, scanned := scanCatalogForReadBacks(t, pat)
			if scanned == 0 {
				t.Fatal("scanned no catalog files — the glob is wrong, and a lint that reads nothing passes everything")
			}
			if len(findings) == 0 {
				return
			}
			sort.Strings(findings)
			t.Errorf("a %s total is read back after a %s change, which reads zero when the change "+
				"paused on a CR 616 prompt (#793 / #807).\n%s; if this read is fine, add it to the "+
				"pattern's allowlist with the reason.\n  %s",
				pat.name, pat.name, pat.advice, strings.Join(findings, "\n  "))
		})
	}
}

// scanCatalogForReadBacks parses every non-test file in the catalog
// package and reports pat's findings, plus how many files it read.
func scanCatalogForReadBacks(t *testing.T, pat readBackPattern) ([]string, int) {
	t.Helper()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	fset := token.NewFileSet()
	var findings []string
	scanned := 0

	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		file, err := parser.ParseFile(fset, name, src, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		scanned++
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			findings = append(findings, readBacksIn(fset, name, fn, pat)...)
		}
	}
	return findings, scanned
}

// readBacksIn reports every pat.reads selector inside fn that sits
// after one of pat's calls in the same function and outside that call's
// own arguments. Arguments are excluded because "lose half your life"
// and "deals damage equal to your life total" both have to read the
// total to compute the amount, and that read happens before the change,
// not after it.
func readBacksIn(fset *token.FileSet, file string, fn *ast.FuncDecl, pat readBackPattern) []string {
	var changes []*ast.CallExpr
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if startsTheClock(call, pat) {
			changes = append(changes, call)
		}
		return true
	})
	if len(changes) == 0 {
		return nil
	}

	var out []string
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok || !pat.reads[sel.Sel.Name] {
			return true
		}
		for _, call := range changes {
			if sel.Pos() >= call.Pos() && sel.End() <= call.End() {
				// Inside the call itself — the amount being dealt or
				// paid, read before anything moved.
				return true
			}
		}
		for _, call := range changes {
			if sel.Pos() <= call.End() {
				continue
			}
			where := file + ":" + strconv.Itoa(fset.Position(sel.Pos()).Line)
			if _, ok := pat.allowlist[where]; ok {
				return true
			}
			out = append(out, where+" reads ."+sel.Sel.Name+" after "+
				file+":"+strconv.Itoa(fset.Position(call.Pos()).Line))
			return true
		}
		return true
	})
	return out
}

// startsTheClock reports whether call is one of the pattern's mutating
// entry points — either a direct `g.Name(…)` call or a catalog
// primitive's `Name{…}.Apply(ctx)`, which is how most of the catalog
// reaches the same engine function.
func startsTheClock(call *ast.CallExpr, pat readBackPattern) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if pat.calls[sel.Sel.Name] {
		return true
	}
	if sel.Sel.Name != "Apply" {
		return false
	}
	// Both spellings reach the same place: `GainLife{…}.Apply(ctx)` and
	// the parenthesised `(DealDamage{…}).Apply(ctx)` gofmt insists on
	// when the literal is the whole statement.
	recv := sel.X
	for {
		paren, ok := recv.(*ast.ParenExpr)
		if !ok {
			break
		}
		recv = paren.X
	}
	lit, ok := recv.(*ast.CompositeLit)
	if !ok {
		return false
	}
	ident, ok := lit.Type.(*ast.Ident)
	return ok && pat.primitives[ident.Name]
}

// TestReadBackGuardCatchesTheShape proves the lint above is a lint and
// not a function that returns nil. A catalog with no findings is the
// expected state, which means the scan could rot into a no-op (a wrong
// glob, a renamed entry point) and nothing would say so.
//
// The fixtures are Exsanguinate and Creeping Bloodsucker as they were
// written before #793 and #807: change a total, read it back on the
// next line, gain the difference.
func TestReadBackGuardCatchesTheShape(t *testing.T) {
	const lifeBefore = `package effects

func drain(ctx *Context, opp uuid.UUID, x int) error {
	p := ctx.PlayerByID(opp)
	was := p.Life
	if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), opp, -x); err != nil {
		return err
	}
	lost := was - p.Life
	return GainLife{Player: ctx.Controller(), Amount: lost}.Apply(ctx)
}
`
	const lifeAfter = `package effects

func drain(ctx *Context, opp uuid.UUID, x int) error {
	return ctx.Game.ChangePlayerLifeThenForEffect(ctx.Source(), opp, -x,
		func(g *game.Game, applied int) error {
			return GainLife{Player: ctx.Controller(), Amount: -applied}.Apply(ctx)
		})
}
`
	const damageBefore = `package effects

func bloodsucker(ctx *Context, opp uuid.UUID, n int) error {
	p := ctx.PlayerByID(opp)
	before := p.Life
	if err := ctx.Game.DealDamageToPlayerForEffect(ctx.Source(), opp, n); err != nil {
		return err
	}
	dealt := before - p.Life
	return GainLife{Player: ctx.Controller(), Amount: dealt}.Apply(ctx)
}
`
	const damageViaPrimitiveBefore = `package effects

func fight(ctx *Context, target uuid.UUID, n int) error {
	c, _ := ctx.Game.LookupCardForEffect(target)
	if err := (DealDamage{Source: ctx.Source(), Target: target, Amount: n}).Apply(ctx); err != nil {
		return err
	}
	return GainLife{Player: ctx.Controller(), Amount: c.DamageMarked}.Apply(ctx)
}
`
	const damageAfter = `package effects

func bloodsucker(ctx *Context, opps []uuid.UUID, n int) error {
	return ctx.Game.DealDamageEachThenForEffect(ctx.Source(), opps, n,
		func(g *game.Game, dealt int) error {
			return GainLife{Player: ctx.Controller(), Amount: dealt}.Apply(ctx)
		})
}
`
	for _, tc := range []struct {
		name string
		pat  readBackPattern
		src  string
		want int
	}{
		{"the pre-#793 life read-back", lifeReadBacks, lifeBefore, 1},
		{"the life continuation", lifeReadBacks, lifeAfter, 0},
		{"the pre-#807 damage read-back", damageReadBacks, damageBefore, 1},
		{"a damage read-back through the primitive", damageReadBacks, damageViaPrimitiveBefore, 1},
		{"the damage continuation", damageReadBacks, damageAfter, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "fixture.go", tc.src, 0)
			if err != nil {
				t.Fatalf("parse fixture: %v", err)
			}
			got := 0
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok {
					got += len(readBacksIn(fset, "fixture.go", fn, tc.pat))
				}
			}
			if got != tc.want {
				t.Errorf("findings = %d, want %d — the lint has stopped seeing the shape it exists for", got, tc.want)
			}
		})
	}
}
