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
// `go vet`.
//
// Every life change runs the CR 614 window (#482), so every life change
// can PAUSE on a CR 616 ordering prompt when two different life
// replacements apply to it. A card that changes a player's life and
// then reads that player's life total back to find out how much moved
// — "each opponent loses X life, you gain life equal to the life lost
// this way" — reads it before the prompt is answered and sees a total
// that has not moved. It compiles, it passes every test played on a
// table with fewer than two life replacements, and it is wrong.
//
// The right shape is the continuation: ChangePlayerLifeThenForEffect
// for one player, LoseLifeEachThenForEffect for several. The engine
// runs `then` with the amount that ACTUALLY moved, whether that is
// immediately or after the prompt.
//
// So this test reads the catalog's own source and fails on the shape
// that was wrong: a `.Life` read that comes after a life change in the
// same function. It is a source scan rather than a behaviour test on
// purpose — the behaviour test needs two life replacements on the
// table to catch anything, which is exactly the board nobody writes a
// test for.
//
// FALSE POSITIVES ARE EXPECTED AND CHEAP. `.Life` is also a field on
// several cost structs, and a loop that reads one player's life and
// then changes another's is fine. Add an entry to lifeReadAllowlist
// with a reason; that is the interruption this test is for.

// lifeChangeCalls are the engine entry points that move a life total
// and may pause. The *Then* forms and the cost form are deliberately
// absent: they are the answer, not the problem.
var lifeChangeCalls = map[string]bool{
	"ChangePlayerLifeForEffect": true,
	"ChangePlayerLife":          true,
}

// lifeReadAllowlist exempts one "<file>:<line>" `.Life` read, with the
// reason it is not the bug. Keep every entry justified: an unexplained
// entry is how the lint stops meaning anything.
var lifeReadAllowlist = map[string]string{}

func TestNoLifeReadBackWithoutAContinuation(t *testing.T) {
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
			findings = append(findings, lifeReadBacksIn(fset, name, fn)...)
		}
	}

	if scanned == 0 {
		t.Fatal("scanned no catalog files — the glob is wrong, and a lint that reads nothing passes everything")
	}
	if len(findings) == 0 {
		return
	}
	sort.Strings(findings)
	t.Errorf("a life total is read back after a life change, which reads zero when the change "+
		"paused on a CR 616 prompt (#793).\n"+
		"Use ChangePlayerLifeThenForEffect (one player) or LoseLifeEachThenForEffect (several) and "+
		"take the amount from the continuation; if this read is fine, add it to lifeReadAllowlist "+
		"with the reason.\n  %s", strings.Join(findings, "\n  "))
}

// lifeReadBacksIn reports every `.Life` selector inside fn that sits
// after a life-change call in the same function and outside that
// call's own arguments. Arguments are excluded because "lose half your
// life" has to read the total to compute the amount, and that read
// happens before the change, not after it.
func lifeReadBacksIn(fset *token.FileSet, file string, fn *ast.FuncDecl) []string {
	var changes []*ast.CallExpr
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if ok && lifeChangeCalls[sel.Sel.Name] {
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
		if !ok || sel.Sel.Name != "Life" {
			return true
		}
		for _, call := range changes {
			if sel.Pos() >= call.Pos() && sel.End() <= call.End() {
				// Inside the call itself — the amount being paid.
				return true
			}
		}
		for _, call := range changes {
			if sel.Pos() <= call.End() {
				continue
			}
			where := file + ":" + strconv.Itoa(fset.Position(sel.Pos()).Line)
			if _, ok := lifeReadAllowlist[where]; ok {
				return true
			}
			out = append(out, where+" reads .Life after "+
				file+":"+strconv.Itoa(fset.Position(call.Pos()).Line))
			return true
		}
		return true
	})
	return out
}

// TestLifeReadBackGuardCatchesTheShape proves the lint above is a lint
// and not a function that returns nil. A catalog with no findings is
// the expected state, which means the scan could rot into a no-op
// (a wrong glob, a renamed entry point) and nothing would say so.
//
// The fixture is Exsanguinate as it was written before #793: change a
// life total, read it back on the next line, gain the difference.
func TestLifeReadBackGuardCatchesTheShape(t *testing.T) {
	const before = `package effects

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
	const after = `package effects

func drain(ctx *Context, opp uuid.UUID, x int) error {
	return ctx.Game.ChangePlayerLifeThenForEffect(ctx.Source(), opp, -x,
		func(g *game.Game, applied int) error {
			return GainLife{Player: ctx.Controller(), Amount: -applied}.Apply(ctx)
		})
}
`
	for _, tc := range []struct {
		name string
		src  string
		want int
	}{
		{"the pre-#793 read-back", before, 1},
		{"the continuation", after, 0},
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
					got += len(lifeReadBacksIn(fset, "fixture.go", fn))
				}
			}
			if got != tc.want {
				t.Errorf("findings = %d, want %d — the lint has stopped seeing the shape it exists for", got, tc.want)
			}
		})
	}
}
