package actions

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"testing"
)

// Every constant of type Type in this package is in Types(), and
// nothing else is: the metrics type label is closed over that list, so
// a type missing from it would be counted as "other".
func TestTypesListsEveryTypeConstant(t *testing.T) {
	declared := map[string]bool{}
	fset := token.NewFileSet()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range f.Decls {
			gd, ok := d.(*ast.GenDecl)
			if !ok || gd.Tok != token.CONST {
				continue
			}
			for _, spec := range gd.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				id, ok := vs.Type.(*ast.Ident)
				if !ok || id.Name != "Type" {
					continue
				}
				for i, n := range vs.Names {
					lit, ok := vs.Values[i].(*ast.BasicLit)
					if !ok {
						t.Fatalf("%s: a Type constant whose value is not a string literal", n.Name)
					}
					v, err := strconv.Unquote(lit.Value)
					if err != nil {
						t.Fatalf("%s: %v", n.Name, err)
					}
					declared[v] = true
				}
			}
		}
	}
	if len(declared) == 0 {
		t.Fatal("found no Type constants; the parser is not reading this package")
	}
	listed := map[string]bool{}
	for _, ty := range Types() {
		if listed[string(ty)] {
			t.Errorf("Types() lists %q twice", ty)
		}
		listed[string(ty)] = true
	}
	for v := range declared {
		if !listed[v] {
			t.Errorf("Type constant %q is not in Types() (types.go)", v)
		}
	}
	for v := range listed {
		if !declared[v] {
			t.Errorf("Types() lists %q, which is not a declared Type constant", v)
		}
	}
}
