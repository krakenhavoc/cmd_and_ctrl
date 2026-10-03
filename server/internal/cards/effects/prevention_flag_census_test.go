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

// prevention_flag_census_test.go — ADR 0107 §5 decision 1 (#1853).
//
// CR 615.12 reads one bit: is this replacement a PREVENTION effect? A
// prevention effect that forgets to say so ignores "damage can't be
// prevented" silently — a Skullcrack that a new "prevent all damage"
// card shrugs off — and nothing else in the tree would ever go red.
//
// So this test reads every ReplacementEffect literal in the catalog and
// in the engine, straight out of the source (the mechanism
// replacement_kind_gate_test.go uses), and fails one that watches damage
// and whose Replace cancels the event or rewrites its amount, unless it
// DECLARES what it is: `Prevention:` (true for a prevention effect,
// false for one that is not, such as "deals damage equal to its power
// instead") or `RedirectsDamage: true`. A Replace that redirects the
// damage — through game.RedirectDamageEventForEffect, the one primitive
// (ADR 0108 §9) — must declare RedirectsDamage, and a Replace that writes
// DamageTarget by hand fails outright: the primitive rewrites the
// recipient and the damage tail together, and does nothing when CR 614.9
// says so.
//
// What it cannot see: a Replace that hands off to a function in another
// file. The engine's scoped shields do that, and declare the flag from
// their kind (scopedKindPrevents) with their own tests.

func TestDamageReplacementsDeclareWhetherTheyPrevent(t *testing.T) {
	var problems []string
	for _, dir := range []string{".", filepath.Join("..", "..", "game")} {
		problems = append(problems, damageReplacementProblems(t, dir)...)
	}
	sort.Strings(problems)
	for _, p := range problems {
		t.Error(p)
	}
}

func damageReplacementProblems(t *testing.T, dir string) []string {
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
			if fd, ok := d.(*ast.FuncDecl); ok && fd.Recv == nil {
				funcs[fd.Name.Name] = fd
			}
		}
	}
	var problems []string
	check := func(lit *ast.CompositeLit) {
		keys := map[string]ast.Expr{}
		for _, el := range lit.Elts {
			if kv, ok := el.(*ast.KeyValueExpr); ok {
				if id, ok := kv.Key.(*ast.Ident); ok {
					keys[id.Name] = kv.Value
				}
			}
		}
		if !mentions(keys["Watches"], "EventDealDamage") {
			return
		}
		body := replaceBody(keys["Replace"], funcs)
		if body == nil {
			return
		}
		changesAmount, redirects, writesTarget := inspectReplace(body)
		where := fset.Position(lit.Pos()).String()
		if writesTarget {
			problems = append(problems, where+": a damage replacement writes DamageTarget by hand; redirect through RedirectDamageEventForEffect (ADR 0108 §9)")
		}
		_, declaresPrevention := keys["Prevention"]
		_, declaresRedirect := keys["RedirectsDamage"]
		if redirects && !declaresRedirect {
			problems = append(problems, where+": a damage replacement that redirects the damage must declare RedirectsDamage: true")
		}
		if changesAmount && !declaresPrevention && !declaresRedirect {
			problems = append(problems, where+": a damage replacement that cancels or rewrites the amount must declare Prevention (CR 615.1a, 615.12) or RedirectsDamage")
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
	return problems
}

func isReplacementEffectType(e ast.Expr) bool {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name == "ReplacementEffect"
	case *ast.SelectorExpr:
		return x.Sel.Name == "ReplacementEffect"
	}
	return false
}

func isReplacementEffectSlice(e ast.Expr) bool {
	a, ok := e.(*ast.ArrayType)
	return ok && isReplacementEffectType(a.Elt)
}

// mentions reports whether an expression names the identifier anywhere.
func mentions(e ast.Expr, name string) bool {
	if e == nil {
		return false
	}
	found := false
	ast.Inspect(e, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && id.Name == name {
			found = true
		}
		return !found
	})
	return found
}

// replaceBody is the body of a Replace field: a func literal, or a
// package-level func named by an identifier.
func replaceBody(e ast.Expr, funcs map[string]*ast.FuncDecl) *ast.BlockStmt {
	switch x := e.(type) {
	case *ast.FuncLit:
		return x.Body
	case *ast.Ident:
		if fd := funcs[x.Name]; fd != nil {
			return fd.Body
		}
	}
	return nil
}

// inspectReplace reports whether a Replace body cancels the event or
// writes its amount other than by multiplying or adding (a doubler or a
// +2 is plainly not a prevention effect), whether it redirects the damage
// through the primitive, and whether it writes the target by hand.
func inspectReplace(body *ast.BlockStmt) (changesAmount, redirects, writesTarget bool) {
	ast.Inspect(body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Cancel" && len(x.Args) == 0 {
				changesAmount = true
			}
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "RedirectDamageEventForEffect" {
				redirects = true
			}
		case *ast.AssignStmt:
			for _, lhs := range x.Lhs {
				sel, ok := lhs.(*ast.SelectorExpr)
				if !ok {
					continue
				}
				switch sel.Sel.Name {
				case "DamageAmount":
					if x.Tok != token.MUL_ASSIGN && x.Tok != token.ADD_ASSIGN {
						changesAmount = true
					}
				case "DamageTarget":
					writesTarget = true
				}
			}
		case *ast.IncDecStmt:
			if sel, ok := x.X.(*ast.SelectorExpr); ok && sel.Sel.Name == "DamageAmount" && x.Tok == token.DEC {
				changesAmount = true
			}
		}
		return true
	})
	return changesAmount, redirects, writesTarget
}
