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
				// The second half of the message is the one that
				// matters since ADR 0083: a token that PRINTS an
				// ability is not a row here at all, and three keys
				// that used to be rows (the Pest, the 0/1 black
				// Wizard, Fable's Goblin Shaman) were deleted rather
				// than left as textless versions of tokens that have
				// never been printed textless. Adding the row back is
				// the wrong fix and this says so.
				t.Errorf("%s:%d: TokenCard(%q) names no token in tokens_table.go. "+
					"If this token PRINTS an ability, it is a tokenTemplate in the "+
					"catalog and not a row here — call its constructor (PestToken(), "+
					"TreasureToken(), …) instead of adding a textless row.",
					f, fset.Position(lit.Pos()).Line, key)
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
		if got, want := c.Colors, colorsNamedInKey(k); !sameColors(got, want) {
			t.Errorf("%q: key names %v, row declares %v — the key reads like the printed text, "+
				"so the two cannot disagree (#1127)", k, want, got)
		}
	}
}

// colorNames maps the colour words a key may use to their letters.
// "colorless" maps to nothing, which is the empty Colors a colourless
// token carries.
var colorNames = map[string]string{
	"white": "W", "blue": "U", "black": "B", "red": "R", "green": "G",
}

// colorsNamedInKey reads the colour clause out of a table key — the
// words between the printed P/T and the creature type, joined by
// "and": "1/1 white Soldier", "1/1 black and green Pest",
// "10/10 colorless Eldrazi". A key with no P/T prefix (the Munitions
// artifact) names no colour and answers nil.
//
// #1127: a template's Colors is a CHARACTERISTIC the rules read, and
// the key is how a reviewer checks it at a glance. Nothing hermetic
// can say which colour is right — that is the printed token, and
// TestEveryTokenTemplateMatchesAPrintedToken asks the dump — but the
// two halves of the row must at least agree with each other, and a
// row edited on one side only is the way the 21 colourless templates
// drifted in the first place.
func colorsNamedInKey(key string) []string {
	rest := key
	slash := strings.Index(rest, "/")
	if slash < 0 {
		return nil
	}
	sp := strings.Index(rest, " ")
	if sp < 0 || sp < slash {
		return nil
	}
	words := strings.Fields(rest[sp+1:])
	var out []string
	for i := 0; i < len(words); i++ {
		if words[i] == "and" && i > 0 && len(out) > 0 {
			continue
		}
		if words[i] == "colorless" && i == 0 {
			return nil
		}
		letter, ok := colorNames[words[i]]
		if !ok {
			break
		}
		out = append(out, letter)
		// A colour word is only part of the clause while the words
		// keep alternating colour / "and".
		if i+1 < len(words) && words[i+1] != "and" {
			break
		}
	}
	return out
}

func sameColors(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
