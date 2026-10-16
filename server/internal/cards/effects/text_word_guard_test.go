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
//
// Slice 2 (#3872) adds the helpers that take the word as an ARGUMENT
// and judge it as given (textWordHelperArgs: ControlsA, the land
// counts, b18ControlsPermanentOfColor…): a literal there is a finding
// too, and the caller passes the word through its text instead
// (ControlsA(ctx.TextSubtype("Faerie")), source.TextSubtype("Island")).
// The helpers that read the word through the text THEMSELVES — the
// predicate builders a lending walk evaluates (IsLandWithSubtype,
// MatchLandSubtype, MatchColor…), the statics that read their source
// (SetsBasicLandType, EachLandIsAlso), the effects that read their
// context (TargetLandBecomes…, the prevention shields' QueryColors) —
// are textWordDoors: a literal is how a card file names its word to
// them, and text_word_helpers_test.go holds them to following a change.

// textWordLiteralAllowed are the files whose literal is correctly
// untranslated, each with the reason. Keep it short: a reason that is
// "not migrated yet" belongs in a migration, not here.
var textWordLiteralAllowed = map[string]string{
	// "You may cast this card from your graveyard as long as you control
	// a black or green permanent": the clause works from the graveyard,
	// where no text change reaches (CR 400.7; a change lives on a
	// permanent or a spell).
	"marang_river_prowler.go": "graveyard-only clause",

	// An alternative cost's condition ("If you control a Plains, you may
	// tap an untapped creature you control rather than pay this spell's
	// mana cost") is judged as the card is cast from a hand (CR 601.2b),
	// before anything can change its text: a change lives on a permanent
	// or a spell already on the stack.
	"angelic_favor.go":  "alternative cost, judged as cast",
	"lashknife.go":      "alternative cost, judged as cast",
	"orims_cure.go":     "alternative cost, judged as cast",
	"ramosian_rally.go": "alternative cost, judged as cast",
	"sivvis_ruse.go":    "alternative cost, judged as cast",
	"sivvis_valor.go":   "alternative cost, judged as cast",
	"snuff_out.go":      "alternative cost, judged as cast",

	// "This land enters tapped unless you control …": asked as the land
	// enters, a new object no text change has reached (CR 400.7, 614.12).
	"dwarven_mine.go":      "enters-tapped condition, judged as it enters",
	"gingerbread_cabin.go": "enters-tapped condition, judged as it enters",
	"idyllic_grange.go":    "enters-tapped condition, judged as it enters",
}

// textWordHelperArgs are effects-package helpers that take a colour or
// land-type word as an argument and judge it AS GIVEN, with the index of
// each such argument. The caller reads the word through its text first.
var textWordHelperArgs = map[string][]int{
	"ControlsA":                         {0},
	"AnOpponentControlsAAndYouControlA": {0, 1},
	"b08LandsWithSubtypeControlled":     {2},
	"b12OtherMountainsControlled":       {3},
	"b18ControlsPermanentOfColor":       {2},
	"b33IslandsControlled":              {2},
	"leftAsColor":                       {2},
}

// textWordDoors are the helpers that read a word argument through the
// text of the object whose clause it is, so a literal passed to them
// follows a change. text_word_helpers_test.go holds them to that; this
// list is the scan's control that they are still in use.
var textWordDoors = []string{
	"IsLandWithSubtype", "IsBasicLandWithSubtype", "MatchLandSubtype", "MatchColor",
	"MatchCreatureSubtype", "SetsBasicLandType", "EachLandIsAlso", "TargetLandBecomesUntilEOT",
	"TargetLandBecomesWhileSourceRemains", "TargetLandBecomesIndefinitely",
	"TargetLandBecomesUntilItsControllersNextTurn", "QueryColors",
}

// textWordReaders are the calls whose first (or, for devotionTo, last)
// argument is a word of the card's text.
var textWordReaders = map[string]bool{
	"HasColor":   true, // Card.HasColor(letter)
	"HasSubtype": true, // Card.HasSubtype(sub), a method; the effects-package HasSubtype predicate reads through the lens
	"devotionTo": true, // devotionTo(g, controller, letter)
}

// isTextWordArgument is isTextWordLiteral in any case: the helpers in
// textWordHelperArgs compare case-insensitively, so "swamp" names the
// word as surely as "Swamp" does.
func isTextWordArgument(lit string) bool {
	return isTextWordLiteral(lit) || game.IsBasicLandType(lit)
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
			if id, ok := call.Fun.(*ast.Ident); ok {
				seen[id.Name]++
				for _, i := range textWordHelperArgs[id.Name] {
					if i >= len(call.Args) {
						continue
					}
					lit, ok := call.Args[i].(*ast.BasicLit)
					if !ok || lit.Kind != token.STRING {
						continue
					}
					word, err := strconv.Unquote(lit.Value)
					if err != nil || !isTextWordArgument(word) {
						continue
					}
					if _, ok := textWordLiteralAllowed[file]; ok {
						usedAllowance[file] = true
						continue
					}
					findings = append(findings, file+":"+strconv.Itoa(fset.Position(call.Pos()).Line)+" "+id.Name+"("+lit.Value+")")
				}
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
	for name := range textWordHelperArgs {
		if seen[name] == 0 {
			t.Errorf("no catalog file calls %s — drop it from textWordHelperArgs", name)
		}
	}
	for _, name := range textWordDoors {
		if seen[name] == 0 {
			t.Errorf("no catalog file calls %s — drop it from textWordDoors", name)
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
