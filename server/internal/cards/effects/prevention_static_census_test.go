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

// prevention_static_census_test.go — ADR 0108 §8 decision 3 (#1906).
//
// A prevention static's additional effect is owed with what its Replace
// PREVENTED, and the apply loop learns that by measuring the event before
// and after Replace (game/prevention_then.go). That measure is the whole
// truth only if Replace does nothing but change the event: a Replace that
// also put a counter on something, or dealt damage, would be an
// additional effect the measure cannot see — and one CR 615.12 skips,
// because under damage that can't be prevented Replace is not run at all.
// The additional effect belongs in Then, where CR 615.12 still runs it.
//
// So this test reads every ReplacementEffect literal that declares
// `Prevention: true`, in the catalog and in the engine, and fails one
// whose Replace calls anything but a method of the event it is handed (or
// min / max / len), or assigns to anything but the event's fields and its
// own locals.
//
// What it cannot see: a Replace that hands off to a function in another
// file, as the engine's scoped shields do (their Prevention is computed
// from their kind, never the literal true, and their follow-ups are owed
// by the kind's own applier, with its own tests).
func TestPreventionStaticsOnlyPrevent(t *testing.T) {
	var problems []string
	for _, dir := range []string{".", filepath.Join("..", "..", "game")} {
		problems = append(problems, preventionReplaceProblems(t, dir)...)
	}
	sort.Strings(problems)
	for _, p := range problems {
		t.Error(p)
	}
}

func preventionReplaceProblems(t *testing.T, dir string) []string {
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
		if v, ok := keys["Prevention"].(*ast.Ident); !ok || v.Name != "true" {
			return
		}
		ft, body := replaceFunc(keys["Replace"], funcs)
		if body == nil {
			return
		}
		where := fset.Position(lit.Pos()).String()
		for _, p := range replaceSideEffects(ft, body) {
			problems = append(problems, where+": a prevention effect's Replace may only change the event; "+p+
				" (an additional effect goes in Then, ADR 0108 §8, where CR 615.12 still runs it)")
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

// replaceFunc is the Replace field's function, its signature and its
// body: a literal, or a package-level function named by an identifier.
func replaceFunc(e ast.Expr, funcs map[string]*ast.FuncDecl) (*ast.FuncType, *ast.BlockStmt) {
	switch x := e.(type) {
	case *ast.FuncLit:
		return x.Type, x.Body
	case *ast.Ident:
		if fd := funcs[x.Name]; fd != nil {
			return fd.Type, fd.Body
		}
	}
	return nil, nil
}

// replaceSideEffects lists what a Replace body does besides changing the
// event it is handed (the signature's first parameter).
func replaceSideEffects(ft *ast.FuncType, body *ast.BlockStmt) []string {
	ev := ""
	if ft.Params != nil && len(ft.Params.List) > 0 && len(ft.Params.List[0].Names) > 0 {
		ev = ft.Params.List[0].Names[0].Name
	}
	onEvent := func(e ast.Expr) bool {
		for {
			switch x := e.(type) {
			case *ast.SelectorExpr:
				e = x.X
			case *ast.IndexExpr:
				e = x.X
			case *ast.Ident:
				return ev != "" && ev != "_" && x.Name == ev
			default:
				return false
			}
		}
	}
	var out []string
	ast.Inspect(body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			switch fn := x.Fun.(type) {
			case *ast.SelectorExpr:
				if onEvent(fn.X) {
					return true
				}
				out = append(out, "it calls "+exprString(fn))
			case *ast.Ident:
				switch fn.Name {
				case "min", "max", "len":
					return true
				}
				out = append(out, "it calls "+fn.Name)
			default:
				out = append(out, "it calls a function value")
			}
		case *ast.AssignStmt:
			for _, lhs := range x.Lhs {
				if _, local := lhs.(*ast.Ident); local || onEvent(lhs) {
					continue
				}
				out = append(out, "it assigns "+exprString(lhs))
			}
		case *ast.IncDecStmt:
			if _, local := x.X.(*ast.Ident); !local && !onEvent(x.X) {
				out = append(out, "it changes "+exprString(x.X))
			}
		case *ast.GoStmt, *ast.DeferStmt, *ast.SendStmt:
			out = append(out, "it starts, defers or sends something")
		}
		return true
	})
	return out
}

// The census itself: what it lets through and what it catches.
func TestPreventionCensusReadsAReplaceBody(t *testing.T) {
	for src, want := range map[string]int{
		`func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error { ev.Cancel(); return nil }`:                                                  0,
		`func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error { n := min(ev.DamageAmount, 1); ev.DamageAmount -= n; return nil }`:           0,
		`func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error { ev.Cancel(); return g.AddCounterForEffect(src.InstanceID, "+1/+1", -1) }`: 1,
		`func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) error { src.Tapped = true; ev.Cancel(); return nil }`:                             1,
		`func(_ *game.ReplacementEvent, _ *game.Game, _ *game.Card) error { return nil }`:                                                                0,
	} {
		e, err := parser.ParseExpr(src)
		if err != nil {
			t.Fatal(err)
		}
		lit := e.(*ast.FuncLit)
		if got := replaceSideEffects(lit.Type, lit.Body); len(got) != want {
			t.Errorf("%s: %d side effects %v, want %d", src, len(got), got, want)
		}
	}
}
