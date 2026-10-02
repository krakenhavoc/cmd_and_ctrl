package effects

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// damage_instance_guard_test.go — ADR 0108 PR 0 (owner decision 4): one
// printed damage instruction is ONE instance of damage, however many
// engine calls the catalog makes to deal it.
//
// The engine stamps each DealDamage…ForEffect call (and each
// DealDamageEachThenForEffect walk) with a DamageInstance of its own, which
// is right for "it deals 2 damage to you. Then it deals 2 damage to you"
// (CR 608.2c: two instructions, two instances). But most of the catalog
// writes "deals 2 damage to each creature" as a loop over DealDamage, and
// that is one instruction and one instance (CR 615.8): a next-damage
// shield on "you and/or creatures you control" must prevent all of it,
// its follow-up must run once with the total (CR 615.5), and ADR 0108's
// later readers — divide_shield, the Phantoms' one counter per
// application, Phyrexian Unlife's life check — all read the same unit.
// A loop that deals damage outside a game.DamageInstanceForEffect scope
// silently splits its instruction into one instance per recipient.
//
// So this test reads the catalog's source and fails on a damage call
// inside a for/range loop that is not lexically inside a
// DamageInstanceForEffect scope. A loop body that RETURNS the damage call
// (`for _, t := range ctx.LegalTargets() { return DealDamage{…}.Apply(ctx) }`)
// deals it at most once and is not a finding.
//
// What it cannot see: one instruction written as two calls with no loop
// (a fight, "X damage to target creature and 1 damage to each other
// creature"). Those are wrapped by hand (b10Fight, Fear, Fire, Foes!,
// Chandra's Ignition, Earthquake); a new one needs the same.
func TestDamageLoopsAreOneInstance(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(files)
	fset := token.NewFileSet()
	var findings []string
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		af, err := parser.ParseFile(fset, f, src, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, pos := range unscopedDamageLoops(af) {
			findings = append(findings, fset.Position(pos).String())
		}
	}
	if len(findings) > 0 {
		t.Errorf(`these damage calls run in a loop outside a damage-instance scope:

  %s

Each printed instruction is one instance of damage (CR 615.8, ADR 0108 PR 0).
Wrap the loop — and anything else the same sentence deals — in

	ctx.Game.DamageInstanceForEffect(func() error { … })

or deal it with DealDamageEachThenForEffect.`, strings.Join(findings, "\n  "))
	}
}

var damageEntryPoint = regexp.MustCompile(`^(DealDamage[A-Za-z]*ForEffect|DealMarkedDamageForEffect)$`)

// isCatalogDamageCall reports whether n deals damage: an engine entry
// point, or the DealDamage primitive's Apply.
func isCatalogDamageCall(n ast.Node) bool {
	c, ok := n.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := c.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if damageEntryPoint.MatchString(sel.Sel.Name) {
		return true
	}
	if sel.Sel.Name != "Apply" {
		return false
	}
	x := sel.X
	if p, ok := x.(*ast.ParenExpr); ok {
		x = p.X
	}
	cl, ok := x.(*ast.CompositeLit)
	if !ok {
		return false
	}
	id, ok := cl.Type.(*ast.Ident)
	return ok && id.Name == "DealDamage"
}

// isDamageScope reports whether call opens a damage-instance scope.
func isDamageScope(n ast.Node) bool {
	c, ok := n.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := c.Fun.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "DamageInstanceForEffect"
}

// unscopedDamageLoops lists the damage calls in a loop, outside a scope.
func unscopedDamageLoops(af *ast.File) []token.Pos {
	var out []token.Pos
	var walk func(n ast.Node, inLoop, inScope bool)
	walk = func(n ast.Node, inLoop, inScope bool) {
		ast.Inspect(n, func(m ast.Node) bool {
			if m == nil || m == n {
				return true
			}
			switch t := m.(type) {
			case *ast.FuncLit:
				// A function literal is a new body: a loop around its
				// definition does not run its calls, but a scope it is
				// handed to does.
				walk(t.Body, false, inScope)
				return false
			case *ast.ForStmt:
				walk(t.Body, true, inScope)
				return false
			case *ast.RangeStmt:
				walk(t.Body, true, inScope)
				return false
			case *ast.ReturnStmt:
				if len(t.Results) == 1 && isCatalogDamageCall(t.Results[0]) {
					return false
				}
			}
			if isDamageScope(m) {
				for _, a := range m.(*ast.CallExpr).Args {
					walk(a, inLoop, true)
				}
				return false
			}
			if inLoop && !inScope && isCatalogDamageCall(m) {
				out = append(out, m.Pos())
			}
			return true
		})
	}
	for _, d := range af.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Body != nil {
			walk(fd.Body, false, false)
		}
	}
	return out
}
