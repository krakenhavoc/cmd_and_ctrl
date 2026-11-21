package effectstest

import (
	"go/ast"
)

// exprString renders a small expression for the scans above.
func ExprString(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.SelectorExpr:
		return ExprString(x.X) + "." + x.Sel.Name
	case *ast.ArrayType:
		return "[]" + ExprString(x.Elt)
	case *ast.StarExpr:
		return "*" + ExprString(x.X)
	case *ast.CallExpr:
		return ExprString(x.Fun) + "()"
	}
	return ""
}
