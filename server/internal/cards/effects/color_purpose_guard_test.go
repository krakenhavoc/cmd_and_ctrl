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

// color_purpose_guard_test.go is #780's lint, in the style of #806's
// (life_continuation_guard_test.go) and #810's
// (x_matters_guard_test.go): a source scan of the catalog for a
// declaration that has to be there and that nothing else can derive.
//
// CR 105.4 makes all five colours a legal answer to "choose a color",
// so the prompt itself says nothing about which one the card wants. A
// chooser with no other information names its own main colour — right
// for Coldsteel Heart, and a self-inflicted board wipe for Wash Out.
// `game.ColorPurpose` is the card's one-word answer to "what happens to
// the colour I name", and the bot's policy switches on it
// (aiseat/heuristic/choices.go, colorChoiceValue).
//
// The declaration is a required PARAMETER of each builder, so the
// compiler already refuses a prompt that passes nothing. What the
// compiler cannot refuse is a prompt that passes the zero value, or a
// string literal, or a variable nobody can trace — and the zero value
// of a string type is the "declares nothing" case the policy
// deliberately keeps on the old rule. So this scan insists the argument
// is one of the named constants, spelled out at the call site where a
// reviewer reading the card will see it.

// colorPromptBuilders are the three doors to a `choose_color` prompt.
// Every one takes the purpose as its FIRST argument.
var colorPromptBuilders = map[string]bool{
	"ChooseColorAsEnters":          true,
	"ChooseColorOtherThanAsEnters": true,
	"ChooseColorThen":              true,
}

// colorPurposeNames is the declared vocabulary, by the identifier a
// card file spells: `game.ColorForMana` and friends. Built from
// game.AllColorPurposes so a purpose added to the engine cannot be
// missed here.
func colorPurposeNames() map[string]bool {
	byValue := map[game.ColorPurpose]string{
		game.ColorForMana:       "ColorForMana",
		game.ColorForBenefit:    "ColorForBenefit",
		game.ColorForHarm:       "ColorForHarm",
		game.ColorForFilter:     "ColorForFilter",
		game.ColorForProtection: "ColorForProtection",
	}
	out := map[string]bool{}
	for _, p := range game.AllColorPurposes {
		name, ok := byValue[p]
		if !ok {
			// A purpose was added to game.AllColorPurposes and not to
			// the table above; fail loudly rather than pass vacuously.
			continue
		}
		out[name] = true
	}
	return out
}

func TestEveryChooseColorPromptDeclaresAPurpose(t *testing.T) {
	if len(colorPurposeNames()) != len(game.AllColorPurposes) {
		t.Fatalf("colorPurposeNames maps %d of %d game.AllColorPurposes — add the new one to the table",
			len(colorPurposeNames()), len(game.AllColorPurposes))
	}

	fset := token.NewFileSet()
	files := parseCatalogSources(t, fset)
	if len(files) == 0 {
		t.Fatal("scanned no catalog files — the glob is wrong, and a lint that reads nothing passes everything")
	}

	names := colorPurposeNames()
	seen := map[string]int{}
	var findings []string

	for file, tree := range files {
		if file == "color_choice.go" {
			// The builders themselves: they take the purpose as a
			// parameter and pass it on. Nothing to declare.
			continue
		}
		ast.Inspect(tree, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			fn, ok := call.Fun.(*ast.Ident)
			if !ok || !colorPromptBuilders[fn.Name] {
				return true
			}
			seen[fn.Name]++
			where := file + ":" + strconv.Itoa(fset.Position(call.Pos()).Line)
			if len(call.Args) == 0 {
				findings = append(findings, where+" "+fn.Name+" is called with no arguments at all")
				return true
			}
			sel, ok := call.Args[0].(*ast.SelectorExpr)
			if !ok || !names[sel.Sel.Name] {
				findings = append(findings,
					where+" "+fn.Name+" does not declare a purpose: its first argument is "+
						exprText(call.Args[0])+", not one of game.ColorFor…")
			}
			return true
		})
	}

	// The control: a lint that stopped matching anything would pass
	// every card forever. Every builder must be exercised by the
	// catalog, or the scan has rotted against a rename.
	for name := range colorPromptBuilders {
		if seen[name] == 0 {
			t.Errorf("no catalog card calls %s — the scan is matching a name that no longer exists", name)
		}
	}

	if len(findings) == 0 {
		return
	}
	sort.Strings(findings)
	t.Errorf("%d choose_color prompt(s) declare no purpose (#780).\n"+
		"Every prompt passes one of %v as its first argument, so the bot's policy "+
		"(aiseat/heuristic/choices.go, colorChoiceValue) knows whether the colour it names is "+
		"about to be helped, harmed, filtered or protected against — CR 105.4 makes all five "+
		"legal, so nothing else can say.\n  %s",
		len(findings), game.AllColorPurposes, strings.Join(findings, "\n  "))
}

// TestColorPromptsGoThroughTheBuilders keeps the guard above
// load-bearing: a card file that queued the engine prompt directly
// would carry no purpose and the scan would never see it.
func TestColorPromptsGoThroughTheBuilders(t *testing.T) {
	fset := token.NewFileSet()
	files := parseCatalogSources(t, fset)
	var findings []string
	for file, tree := range files {
		if file == "color_choice.go" {
			continue
		}
		ast.Inspect(tree, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			switch sel.Sel.Name {
			case "QueueColorChoiceForEffect", "QueueColorChoiceThenForEffect":
				findings = append(findings,
					file+":"+strconv.Itoa(fset.Position(sel.Pos()).Line)+" calls "+sel.Sel.Name+" directly")
			}
			return true
		})
	}
	if len(findings) > 0 {
		sort.Strings(findings)
		t.Errorf("%d catalog site(s) queue a colour prompt without going through color_choice.go's builders, "+
			"which is how a prompt gets on the wire with no purpose (#780):\n  %s",
			len(findings), strings.Join(findings, "\n  "))
	}
}

// exprText renders a small expression for an error message. Anything
// that is not a bare identifier, a selector or a literal reads as its
// Go type, which is enough to find it.
func exprText(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		return exprText(v.X) + "." + v.Sel.Name
	case *ast.BasicLit:
		return v.Value
	}
	return "an expression"
}
