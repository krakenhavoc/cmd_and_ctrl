package effects

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

// tokens_table_test.go — #581. TokenCard panics on an unknown key, so
// every key written in the package is checked here, where a typo is
// a red test rather than a mid-game panic.

func TestEveryTokenKeyResolves(t *testing.T) {
	fset := token.NewFileSet()
	files, err := filepath.Glob("*.go")
	if err != nil || len(files) == 0 {
		t.Fatalf("glob: %v (%d files)", err, len(files))
	}
	seen := 0
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		af, err := parser.ParseFile(fset, f, src, 0)
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		ast.Inspect(af, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if id, ok := call.Fun.(*ast.Ident); !ok || id.Name != "TokenCard" || len(call.Args) != 1 {
				return true
			}
			lit, ok := call.Args[0].(*ast.BasicLit)
			if !ok {
				return true // a key built at runtime is that caller's own risk
			}
			key, _ := strconv.Unquote(lit.Value)
			if _, ok := tokenTable[key]; !ok {
				t.Errorf("%s:%d: TokenCard(%q) names no token in tokens_table.go", f, fset.Position(lit.Pos()).Line, key)
			}
			seen++
			return true
		})
	}
	if seen == 0 {
		t.Fatal("no TokenCard(...) literals found; the scan is broken")
	}
}

func TestTokenCardReturnsFreshCopies(t *testing.T) {
	a := TokenCard("1/1 black Bat with flying")
	a.Keywords = append(a.Keywords, "haste")
	a.Colors = append(a.Colors, "R")
	b := TokenCard("1/1 black Bat with flying")
	if len(b.Keywords) != 1 || len(b.Colors) != 1 {
		t.Errorf("a caller's append leaked into the next copy: %+v", b)
	}
	if b.Name != "Bat" || !strings.Contains(b.TypeLine, "Bat") || b.Power != 1 || b.Toughness != 1 {
		t.Errorf("unexpected template: %+v", b)
	}
	defer func() {
		if recover() == nil {
			t.Error("an unknown key must panic at the call, not return a zero card")
		}
	}()
	// Built at runtime on purpose: the literal scan above would flag it.
	bogus := "9/9 plaid " + "Wombat"
	_ = TokenCard(bogus)
}

func TestTokenTableRowsAreWellFormed(t *testing.T) {
	for _, k := range TokenKeys() {
		c := tokenTable[k]
		if c.Name == "" || !strings.HasPrefix(c.TypeLine, "Token ") {
			t.Errorf("%q: name %q, type line %q — a token template needs both", k, c.Name, c.TypeLine)
		}
		if strings.Contains(c.TypeLine, "Creature") && !strings.HasPrefix(k, strconv.Itoa(c.Power)+"/"+strconv.Itoa(c.Toughness)+" ") {
			t.Errorf("%q: key does not start with the printed %d/%d", k, c.Power, c.Toughness)
		}
	}
}
