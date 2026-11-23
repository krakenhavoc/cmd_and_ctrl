package effectstest

import (
	"go/ast"
	"go/token"
	"strings"
	"testing"
)

// parseCatalogSources parses every non-test file in the catalog
// package, keyed by file name.
func ParseCatalogSources(t *testing.T, fset *token.FileSet) map[string]*ast.File {
	t.Helper()
	names, err := FlatGlob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	out := make(map[string]*ast.File, len(names))
	for _, name := range names {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := FlatReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		file, err := FlatParseFile(fset, name, src, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		out[name] = file
	}
	return out
}

// registeredSpecs returns every `Register(Spec{…})` literal in a file.
// A cycle that registers in a loop still builds its Spec as a literal,
// so the scan sees each one once however the call is wrapped.
func RegisteredSpecs(file *ast.File) []*ast.CompositeLit {
	var out []*ast.CompositeLit
	ast.Inspect(file, func(n ast.Node) bool {
		lit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		if ident, ok := lit.Type.(*ast.Ident); ok && ident.Name == "Spec" {
			out = append(out, lit)
		}
		return true
	})
	return out
}
