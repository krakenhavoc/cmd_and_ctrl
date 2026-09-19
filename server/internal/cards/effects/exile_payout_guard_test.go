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

// exile_payout_guard_test.go is the #911 half of the lint
// life_continuation_guard_test.go is the #793 / #807 half of: a source
// scan for a mistake `go vet` cannot see and a behaviour test will not
// catch on an ordinary board.
//
// THE SHAPE. An exile opens the CR 614 window, so it can be cancelled
// ("cards in graveyards can't be exiled"), redirected (a commander card
// taking CR 903.9's offer) or merely PAUSED on a prompt. The
// fire-and-forget forms — `ExileTarget` with no `Then`,
// `ExileCardForEffect`, `ExileCardsForEffect` — return nil for all
// three, so a card that pays out on the next line pays out for an exile
// that did not happen. That is what Cling to Dust, Scavenging Ooze and
// Deluge of the Dead did: read the card's type into a local, exile,
// then gate the life / counter / Zombie on the local.
//
// WHAT IS FLAGGED is precisely that: a CONDITION, after a fire-and-
// forget exile, that reads a local the function assigned BEFORE the
// exile. A gate on a pre-exile snapshot is a clause ABOUT THE CARD, and
// a clause about the card has to know whether the card actually went.
//
// WHAT IS NOT FLAGGED, on purpose:
//
//   - a clause that is not a gate. "Exile target creature. Its
//     controller gains life equal to its power" reads a snapshot too,
//     but it happens either way (ADR 0013 §5m item 5), so a pre-exile
//     local used as a VALUE is fine. Only conditions are read.
//   - `if err != nil`. The error is assigned by the exile, not before
//     it, and `err` is excluded by name as well.
//   - an exile with a `Then`. That is the answer, not the problem.
//
// FALSE POSITIVES ARE EXPECTED AND CHEAP, exactly as they are in the
// life guard: add the "<file>:<line>" to the allowlist with the reason
// it is not the bug. An unexplained entry is how a lint stops meaning
// anything.
//
// WHY A SECOND SCANNER rather than a third row in that file's table:
// the life pattern keys on a SELECTOR field read (`p.Life`) anywhere
// after the call, and this one keys on a LOCAL read inside a condition,
// with the call itself qualified by the absence of a struct field. The
// two have no body in common but the glob.

// exilePayoutAllowlist exempts one "<file>:<line>" finding, with the
// reason it is not the bug.
var exilePayoutAllowlist = map[string]string{
	"solitude.go:74": "Solitude's `if power <= 0` is a guard against gaining zero life, not a gate " +
		"on the exile: \"its controller gains life equal to its power\" is an unconditional clause " +
		"about a player, which ADR 0013 §5m item 5 declared ungated and §5t left ungated.",
}

// exileStartsTheClock names the fire-and-forget exiles. The `Then`
// forms and the batch `Then` forms are deliberately absent — they are
// the answer.
var exileStartsTheClock = map[string]bool{
	"ExileCardForEffect":  true,
	"ExileCardsForEffect": true,
}

// exilePrimitives are the catalog structs whose `.Apply(ctx)` reaches
// one of those, but ONLY when the literal declares no `Then`.
var exilePrimitives = map[string]bool{
	"ExileTarget":      true,
	"ExileAllMatching": true,
}

func TestNoExilePayoutWithoutAContinuation(t *testing.T) {
	findings, scanned := scanCatalogForExilePayouts(t)
	if scanned == 0 {
		t.Fatal("scanned no catalog files — the glob is wrong, and a lint that reads nothing passes everything")
	}
	if len(findings) == 0 {
		return
	}
	sort.Strings(findings)
	t.Errorf("a clause is gated on a value read BEFORE a fire-and-forget exile, so it pays out for "+
		"an exile the CR 614 window cancelled, redirected or has not answered yet (#911).\n"+
		"Use ExileThenIfItWas (the \"if it was a <type> card\" family) or ExileTarget.Then and gate "+
		"the clause on `exiled`; if this condition is fine, add it to exilePayoutAllowlist with the "+
		"reason.\n  %s", strings.Join(findings, "\n  "))
}

// scanCatalogForExilePayouts parses every non-test file in the catalog
// package and reports the findings, plus how many files it read.
func scanCatalogForExilePayouts(t *testing.T) ([]string, int) {
	t.Helper()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	fset := token.NewFileSet()
	seen := map[string]bool{}
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
		for _, f := range exilePayoutsIn(fset, name, file) {
			// A nested func literal is walked both on its own and as
			// part of its enclosing declaration, so one mistake can be
			// reported twice. It is one mistake.
			if seen[f] {
				continue
			}
			seen[f] = true
			findings = append(findings, f)
		}
	}
	return findings, scanned
}

// exilePayoutsIn reports every gate-on-a-pre-exile-local inside file.
// It walks function declarations and function literals alike, because
// most of the catalog's bodies are literals on a Spec.
func exilePayoutsIn(fset *token.FileSet, name string, file *ast.File) []string {
	var out []string
	ast.Inspect(file, func(n ast.Node) bool {
		var body *ast.BlockStmt
		switch fn := n.(type) {
		case *ast.FuncDecl:
			body = fn.Body
		case *ast.FuncLit:
			body = fn.Body
		default:
			return true
		}
		if body == nil {
			return true
		}
		out = append(out, exilePayoutsInBody(fset, name, body)...)
		return true
	})
	return out
}

func exilePayoutsInBody(fset *token.FileSet, name string, body *ast.BlockStmt) []string {
	var exiles []*ast.CallExpr
	ast.Inspect(body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok && isFireAndForgetExile(call) {
			exiles = append(exiles, call)
		}
		return true
	})
	if len(exiles) == 0 {
		return nil
	}
	// The locals this function assigned before its FIRST exile. A
	// later one is covered too: anything assigned before the first is
	// assigned before all of them.
	pre := map[string]bool{}
	ast.Inspect(body, func(n ast.Node) bool {
		as, ok := n.(*ast.AssignStmt)
		if !ok || as.End() > exiles[0].Pos() {
			return true
		}
		for _, lhs := range as.Lhs {
			if id, ok := lhs.(*ast.Ident); ok && id.Name != "_" && id.Name != "err" {
				pre[id.Name] = true
			}
		}
		return true
	})
	if len(pre) == 0 {
		return nil
	}

	var out []string
	ast.Inspect(body, func(n ast.Node) bool {
		var cond ast.Expr
		switch s := n.(type) {
		case *ast.IfStmt:
			cond = s.Cond
		case *ast.SwitchStmt:
			cond = s.Tag
		default:
			return true
		}
		if cond == nil {
			return true
		}
		for _, call := range exiles {
			if n.Pos() <= call.End() {
				continue
			}
			read := ""
			ast.Inspect(cond, func(k ast.Node) bool {
				if id, ok := k.(*ast.Ident); ok && pre[id.Name] {
					read = id.Name
				}
				return true
			})
			if read == "" {
				return true
			}
			where := name + ":" + strconv.Itoa(fset.Position(n.Pos()).Line)
			if _, ok := exilePayoutAllowlist[where]; ok {
				return true
			}
			out = append(out, where+" gates on `"+read+"`, read before the exile at "+
				name+":"+strconv.Itoa(fset.Position(call.Pos()).Line))
			return true
		}
		return true
	})
	return out
}

// isFireAndForgetExile reports whether call is an exile that cannot
// tell its caller what landed: a direct `g.ExileCardForEffect(…)`, or
// an `ExileTarget{…}.Apply(ctx)` whose literal declares no `Then`.
func isFireAndForgetExile(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if exileStartsTheClock[sel.Sel.Name] {
		return true
	}
	if sel.Sel.Name != "Apply" {
		return false
	}
	// Both spellings: `ExileTarget{…}.Apply(ctx)` and the
	// parenthesised form gofmt insists on when the literal is the
	// whole statement.
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
	if !ok || !exilePrimitives[ident.Name] {
		return false
	}
	for _, e := range lit.Elts {
		kv, ok := e.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		if k, ok := kv.Key.(*ast.Ident); ok && k.Name == "Then" {
			return false
		}
	}
	return true
}

// TestExilePayoutGuardCatchesTheShape proves the lint is a lint and not
// a function that returns nil. A catalog with no findings is the
// expected state, which means the scan could rot into a no-op — a wrong
// glob, a renamed entry point, a `Then` check that stopped matching —
// and nothing would say so.
//
// The fixtures are Scavenging Ooze before and after #911, plus the two
// shapes that must stay quiet.
func TestExilePayoutGuardCatchesTheShape(t *testing.T) {
	const before = `package effects

func ooze(ctx *Context, id, source uuid.UUID) error {
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	wasCreature := c.IsCreature()
	if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	if !wasCreature {
		return nil
	}
	return AddCounter{Target: source, Kind: "+1/+1", N: 1}.Apply(ctx)
}
`
	const after = `package effects

func ooze(ctx *Context, id, source uuid.UUID) error {
	return ExileThenIfItWas{
		Target: id,
		Was:    WasCreatureCard,
		Then: func(ctx *Context) error {
			return AddCounter{Target: source, Kind: "+1/+1", N: 1}.Apply(ctx)
		},
	}.Apply(ctx)
}
`
	const throughTheEngineCall = `package effects

func ooze(ctx *Context, id uuid.UUID) error {
	c, _ := ctx.Game.LookupCardForEffect(id)
	wasCreature := c.IsCreature()
	if err := ctx.Game.ExileCardForEffect(id); err != nil {
		return err
	}
	if wasCreature {
		return GainLife{Amount: 1}.Apply(ctx)
	}
	return nil
}
`
	const unconditionalClause = `package effects

func plowshares(ctx *Context, id uuid.UUID) error {
	card, _ := ctx.Game.LookupCardForEffect(id)
	power := card.CurrentPower()
	if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	return GainLife{Player: card.Controller, Amount: power}.Apply(ctx)
}
`
	for _, tc := range []struct {
		name string
		src  string
		want int
	}{
		{"the pre-#911 read-back", before, 1},
		{"the same mistake through the engine call", throughTheEngineCall, 1},
		{"the ExileThenIfItWas continuation", after, 0},
		{"an unconditional clause that only READS a snapshot", unconditionalClause, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "fixture.go", tc.src, 0)
			if err != nil {
				t.Fatalf("parse fixture: %v", err)
			}
			if got := len(exilePayoutsIn(fset, "fixture.go", file)); got != tc.want {
				t.Errorf("findings = %d, want %d — the lint has stopped seeing the shape it exists for",
					got, tc.want)
			}
		})
	}
}
