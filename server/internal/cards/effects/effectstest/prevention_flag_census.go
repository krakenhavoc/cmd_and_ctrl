package effectstest

import (
	"go/ast"
)

func IsReplacementEffectType(e ast.Expr) bool {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name == "ReplacementEffect"
	case *ast.SelectorExpr:
		return x.Sel.Name == "ReplacementEffect"
	}
	return false
}

func IsReplacementEffectSlice(e ast.Expr) bool {
	a, ok := e.(*ast.ArrayType)
	return ok && IsReplacementEffectType(a.Elt)
}

// mentions reports whether an expression names the identifier anywhere.
func Mentions(e ast.Expr, name string) bool {
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
