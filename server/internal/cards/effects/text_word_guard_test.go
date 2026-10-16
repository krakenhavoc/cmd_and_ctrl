package effects

import (
	"go/ast"
	"go/token"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// text_word_guard_test.go is #3872's lint, in the style of #780's
// (color_purpose_guard_test.go): a source scan of the catalog for a
// colour word or basic land type written as a Go literal where a
// text-changing effect can't reach it.
//
// Sleight of Mind and Magical Hack change a word of an object's text
// (CR 612.2; ADR 0160's 2026-10-15 amendment). In this engine the words
// are arguments in card files, so a word changes only when the card
// reads it through a door that knows the object's text: TextColor /
// TextSubtype on the source (or ctx.TextColor in an effect), or the
// lent lens (HasColorWord / HasSubtypeWord) inside a predicate the
// targeting walk, a sweep, a search or a cast trigger hands the
// source's text. `c.HasColor("U")` on its own is none of those: Pyroblast
// written that way stays "if it's blue" whatever its text says.
//
// So a literal colour letter or basic land type passed straight to
// Card.HasColor, Card.HasSubtype or devotionTo is a finding, unless the
// file is listed below with the reason the word is out of reach.

// textWordLiteralAllowed are the files whose literal is correctly
// untranslated, each with the reason. Keep it short: a reason that is
// "not migrated yet" belongs in a migration, not here.
var textWordLiteralAllowed = map[string]string{
	// "You may cast this card from your graveyard as long as you control
	// a black or green permanent": the clause works from the graveyard,
	// where no text change reaches (CR 400.7; a change lives on a
	// permanent or a spell).
	"marang_river_prowler.go": "graveyard-only clause",
}

// textWordReaders are the calls whose first (or, for devotionTo, last)
// argument is a word of the card's text.
var textWordReaders = map[string]bool{
	"HasColor":   true, // Card.HasColor(letter)
	"HasSubtype": true, // Card.HasSubtype(sub), a method; the effects-package HasSubtype predicate reads through the lens
	"devotionTo": true, // devotionTo(g, controller, letter)
}

func isTextWordLiteral(lit string) bool {
	switch lit {
	case "W", "U", "B", "R", "G":
		return true
	}
	return game.IsBasicLandType(lit) && strings.ToUpper(lit[:1]) == lit[:1]
}

func TestNoHardCodedTextWordsInCardFiles(t *testing.T) {
	fset := token.NewFileSet()
	files := parseCatalogSources(t, fset)
	if len(files) == 0 {
		t.Fatal("scanned no catalog files — the glob is wrong, and a lint that reads nothing passes everything")
	}

	seen := map[string]int{}
	var findings []string
	usedAllowance := map[string]bool{}
	for file, tree := range files {
		ast.Inspect(tree, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 {
				return true
			}
			var name string
			arg := call.Args[0]
			switch fn := call.Fun.(type) {
			case *ast.SelectorExpr:
				// A method call: c.HasColor("U"), q.Card.HasSubtype("Island").
				// A package-qualified call (game.X) is not a card method.
				if id, ok := fn.X.(*ast.Ident); ok && id.Name == "game" {
					return true
				}
				name = fn.Sel.Name
				if name == "devotionTo" {
					return true
				}
			case *ast.Ident:
				if fn.Name != "devotionTo" {
					return true
				}
				name = fn.Name
				arg = call.Args[len(call.Args)-1]
			default:
				return true
			}
			if !textWordReaders[name] {
				return true
			}
			seen[name]++
			lit, ok := arg.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			word, err := strconv.Unquote(lit.Value)
			if err != nil || !isTextWordLiteral(word) {
				return true
			}
			if _, ok := textWordLiteralAllowed[file]; ok {
				usedAllowance[file] = true
				return true
			}
			findings = append(findings, file+":"+strconv.Itoa(fset.Position(call.Pos()).Line)+" "+name+"("+lit.Value+")")
			return true
		})
	}

	// The control: a scan that stopped matching would pass every card.
	for name := range textWordReaders {
		if seen[name] == 0 {
			t.Errorf("no catalog file calls %s — the scan is matching a name that no longer exists", name)
		}
	}
	for file := range textWordLiteralAllowed {
		if !usedAllowance[file] {
			t.Errorf("%s is allowed a hard-coded text word but no longer has one: drop it from textWordLiteralAllowed", file)
		}
	}

	if len(findings) == 0 {
		return
	}
	sort.Strings(findings)
	t.Errorf("%d colour word(s) or basic land type(s) hard-coded where a text-changing effect can't reach them (#3872, CR 612.2).\n"+
		"Read the word through the object whose text it is: source.TextColor(\"U\") / source.TextSubtype(\"Island\") "+
		"for a static or trigger, ctx.TextColor in an effect, or HasColorWord / HasSubtypeWord inside a predicate a "+
		"lending helper evaluates (ADR 0160's 2026-10-15 amendment). If the word really is out of reach, list the file "+
		"in textWordLiteralAllowed with the reason.\n  %s",
		len(findings), strings.Join(findings, "\n  "))
}
