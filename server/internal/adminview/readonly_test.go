package adminview

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// writeMethods are the database calls that can change something. None
// may appear in this package's non-test code, called or as a value.
var writeMethods = map[string]bool{
	"Exec": true, "ExecContext": true,
	"Begin": true, "BeginTx": true,
	"Prepare": true, "PrepareContext": true,
}

// queryMethods are the reads, and where each takes its query.
var queryMethods = map[string]int{
	"Query": 0, "QueryRow": 0,
	"QueryContext": 1, "QueryRowContext": 1,
}

// readOnlyOffences is every way f could write: a write method named at
// all, or a query that is not a string literal starting with SELECT or
// WITH.
func readOnlyOffences(fset *token.FileSet, f *ast.File) []string {
	var out []string
	ast.Inspect(f, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.SelectorExpr:
			if writeMethods[n.Sel.Name] {
				out = append(out, fset.Position(n.Pos()).String()+": "+n.Sel.Name)
			}
		case *ast.CallExpr:
			sel, ok := n.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			arg, ok := queryMethods[sel.Sel.Name]
			if !ok {
				return true
			}
			if len(n.Args) <= arg {
				out = append(out, fset.Position(n.Pos()).String()+": "+sel.Sel.Name+" with no query")
				return true
			}
			lit, ok := n.Args[arg].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				out = append(out, fset.Position(n.Pos()).String()+": "+sel.Sel.Name+"'s query is not a string literal")
				return true
			}
			q, err := strconv.Unquote(lit.Value)
			if err != nil {
				out = append(out, fset.Position(n.Pos()).String()+": "+err.Error())
				return true
			}
			head := strings.ToUpper(strings.TrimSpace(q))
			if !strings.HasPrefix(head, "SELECT") && !strings.HasPrefix(head, "WITH") {
				out = append(out, fset.Position(n.Pos()).String()+": "+sel.Sel.Name+" query does not start with SELECT or WITH")
			}
		}
		return true
	})
	return out
}

// TestAdminViewStoreOnlyReads is ADR 0124 §8: the admin views' store
// never writes. It fails on a non-test file in this package that names
// Exec, BeginTx or Prepare in any form, or passes a query that is not
// a literal starting with SELECT or WITH.
func TestAdminViewStoreOnlyReads(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	queries := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(".", name), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, o := range readOnlyOffences(fset, f) {
			t.Errorf("%s; the admin views only read (ADR 0124 §5)", o)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if _, ok := queryMethods[sel.Sel.Name]; ok {
						queries++
					}
				}
			}
			return true
		})
	}
	// Control: the walk is finding the store's queries at all.
	if queries < 7 {
		t.Errorf("found only %d queries in the package; the guard is not reading the store", queries)
	}
}

// The rule catches what it exists for, and passes a clean file.
func TestReadOnlyRuleCatchesAWrite(t *testing.T) {
	cases := map[string]struct {
		src  string
		want int
	}{
		"exec": {`package adminview
func f(s *SQLStore) { s.q.ExecContext(ctx, "DELETE FROM users") }`, 1},
		"transaction": {`package adminview
func f(d *db.DB) { tx, _ := d.BeginTx(ctx, nil); _ = tx }`, 1},
		"a method value": {`package adminview
func f(d *db.DB) any { return d.ExecContext }`, 1},
		"an update through a read": {`package adminview
func f(s *SQLStore) { s.q.QueryContext(ctx, "UPDATE users SET display_name = 'x' RETURNING id") }`, 1},
		"a built query": {`package adminview
func f(s *SQLStore, q string) { s.q.QueryRowContext(ctx, "SELECT " + q) }`, 1},
		"clean": {`package adminview
func f(s *SQLStore) { s.q.QueryContext(ctx, ` + "`\n\t WITH a AS (SELECT 1) SELECT * FROM a`" + `); s.q.QueryRowContext(ctx, "select 1") }`, 0},
	}
	for name, tc := range cases {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, name+".go", tc.src, 0)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if got := readOnlyOffences(fset, f); len(got) != tc.want {
			t.Errorf("%s: %d offences %v, want %d", name, len(got), got, tc.want)
		}
	}
}
