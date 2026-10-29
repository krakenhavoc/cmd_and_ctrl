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
//
// #3984 widens the words to creature types (Artificial Evolution), read
// from the engine's vetted CR 205.3m list (game.AllCreatureTypes) in its
// canonical spelling, and adds the colour- and type-setting mods
// (textWordGameMods: "becomes a white and blue Bird"), the turn tallies
// that take a creature type (textWordMethodArgs), and the creature-type
// helpers of the tribal files. text_word_remainder_test.go is its rules
// test.

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
	"bards_company.go":  "alternative cost, judged as cast",
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
	"roads.go":             "enters-tapped condition, judged as it enters",
	// "Enters with a +1/+1 counter for each other Zombie you control" is
	// a replacement read as the creature enters, the same new object.
	"unbreathing_horde.go": "enters-with-counters count, judged as it enters",

	// Clauses that work only from a graveyard (#3984), where no text
	// change reaches, like Marang River Prowler's.
	"gravecrawler.go": "graveyard-only clause",
	"quakebringer.go": "graveyard-only clause",
	// Embalm and eternalize are activated from the graveyard; the token's
	// "except it's a white Zombie" is that card's printed text (CR 400.7).
	"embalm.go": "graveyard-only ability",
	// Submerge's "If an opponent controls a Forest and you control an
	// Island, you may cast this spell without paying its mana cost".
	"submerge.go": "alternative cost, judged as cast",

	// "This spell costs {2} less to cast if you control a Wolf or
	// Werewolf", and a modal spell's "if you control a Wizard as you
	// cast this spell, choose two" (Flame of Anor): judged as the card
	// is cast (CR 601.2b, 601.2f), before anything can change its text.
	"wolfkin_outcast.go": "self cost reduction, judged as cast",
	"modes.go":           "mode count condition, judged as cast",
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

	// #3984: helpers that take a creature type (or a land type) and judge
	// it as given.
	"leftAsSubtype":                    {2},
	"hasSubtype":                       {1},
	"attackerHasSubtype":               {2},
	"attackersOfSubtypeInTheSameBatch": {3},
	"permanentWasCreatureType":         {1},
	"addedSubtypeTypeLine":             {1},
	"retypedTypeLine":                  {fromArg(1)},
	"b05ControlsSubtype":               {2},
	"b10AnotherPermanentWithSubtypeEnteredUnderYourControl": {3},
	"b19ZombieCardsInGraveyard":                             {2},
	"b22OtherCreaturesOfSubtypeControlled":                  {3},
	"b24AttackingCreaturesOfSubtype":                        {2},
	"b24CombatDamageToPlayerByYourCreatureOfSubtypes":       {fromArg(3)},
	"b24SubtypeYouControlDealtDamage":                       {3},
	"b26PermanentsOfSubtypeControlled":                      {2},
	"b27CreatureTypesPlusShapeshifter":                      {1},
	"b28GraveyardHasCreatureCardOfSubtype":                  {2},
	"b29SpiritsControlled":                                  {2},
	"b34GoblinsEnteredUnderYourControlThisTurn":             {2},
	"b38FaeriesYouControl":                                  {2},
	"b42CountCreaturesWithSubtype":                          {1},
	"b43CreaturesOfSubtypeControlled":                       {2},
	"rfCreatureCEnteredLandHasSubtype":                      {2},
}

// fromArg(i) in textWordHelperArgs is "argument i and every one after
// it is a word" (a variadic list).
func fromArg(i int) int { return -1 - i }

// textWordMethodArgs are Game methods (the turn tallies) that take a
// creature type and judge it as given, with the index of the word: the
// caller passes ctx.TextSubtype("Dragon").
var textWordMethodArgs = map[string]int{
	"EnteredWithSubtypeThisTurn":                   1,
	"DiedWithSubtypeThisTurn":                      1,
	"PlayersDealtCombatDamageThisTurnBySubtype":    1,
	"PlayersDealtCombatDamageThisTurnByAnySubtype": 0,
	"DiedPowerWithSubtypeThisTurn":                 0,
}

// textWordGameMods are the game package's colour- and type-setting mods
// (#3984): every argument is a word of the text of the object that
// creates the effect ("becomes a white and blue Bird", "is a black
// Zombie in addition to its other colours and types"). A card file
// passes each through that object's text: game.SetColorsMod(
// ctx.TextColor("W"), ctx.TextColor("U")), game.AddSubtypesMod(
// ctx.TextSubtype("Bird")), source.TextColor in a static.
var textWordGameMods = map[string]bool{
	"SetColorsMod":        true,
	"AddColorsMod":        true,
	"AddSubtypesMod":      true,
	"SetCreatureTypesMod": true,
	"LoseSubtypesMod":     true,
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
	// #3984: creature types, colour- and type-setting statics, costs and
	// mana restrictions.
	"OfCreatureType", "Subtype", "QuerySubtype", "exceptSubtypes",
	"AttachedIsAlso", "SetAttachedColors", "SetAttachedTypes",
	"OtherSpellOfCreatureType", "SpellOfCreatureType",
	"AnotherNontokenCreatureOfTypeEnteredUnderYourControl",
	"thisEnchantmentBecomesACreature", "b24TargetAnyNotSubtype",
	"b16OtherCreaturesYouControlOfSubtype", "b21CreatureOfSubtypeYouControlAttacked",
	"SacrificeSubtype", "RevealX", "CraftWithEachOf", "CraftWithOneOrMoreSubtype",
	"ManaRestrictSubtype", "ManaRestrictColor",
	"anotherEnteredWithSubtypeThisTurn", "b43CreatureOfSubtypeYouControlDealtCombatDamageToPlayer",
	"IsBasicLandOfAnySubtype", "IsLandWithAnySubtype", "CraftWithSubtype", "frodoSauronsBaneIs",
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
	return isTextWordLiteral(lit) || game.IsBasicLandType(lit) || game.IsCreatureType(lit)
}

// isTextWordLiteral is a colour letter, a basic land type, or (#3984) a
// creature type, each in the spelling a card file writes it: the
// creature types are the engine's vetted CR 205.3m vocabulary
// (game.AllCreatureTypes, generated from the catalog's type lines), in
// its canonical spelling, so a word that merely looks like one ("Arcane",
// "Vehicle", "Food", "Equipment" are not creature types) is never a
// finding.
func isTextWordLiteral(lit string) bool {
	switch lit {
	case "W", "U", "B", "R", "G":
		return true
	}
	if game.IsBasicLandType(lit) && strings.ToUpper(lit[:1]) == lit[:1] {
		return true
	}
	canon, ok := game.CanonicalCreatureType(lit)
	return ok && canon == lit
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
			// argFinding records a literal word handed to a helper that
			// judges it as given.
			argFinding := func(name string, arg ast.Expr) {
				lit, ok := arg.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return
				}
				word, err := strconv.Unquote(lit.Value)
				if err != nil || !isTextWordArgument(word) {
					return
				}
				if _, ok := textWordLiteralAllowed[file]; ok {
					usedAllowance[file] = true
					return
				}
				findings = append(findings, file+":"+strconv.Itoa(fset.Position(call.Pos()).Line)+" "+name+"("+lit.Value+")")
			}
			if id, ok := call.Fun.(*ast.Ident); ok {
				seen[id.Name]++
				for _, i := range textWordHelperArgs[id.Name] {
					if i < 0 {
						if from := -1 - i; from < len(call.Args) {
							for _, a := range call.Args[from:] {
								argFinding(id.Name, a)
							}
						}
						continue
					}
					if i < len(call.Args) {
						argFinding(id.Name, call.Args[i])
					}
				}
			}
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if i, ok := textWordMethodArgs[sel.Sel.Name]; ok {
					seen["."+sel.Sel.Name]++
					if i < len(call.Args) {
						argFinding(sel.Sel.Name, call.Args[i])
					}
				}
				if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "game" && textWordGameMods[sel.Sel.Name] {
					seen["game."+sel.Sel.Name]++
					for _, a := range call.Args {
						argFinding("game."+sel.Sel.Name, a)
					}
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
	for name := range textWordMethodArgs {
		if seen["."+name] == 0 {
			t.Errorf("no catalog file calls .%s — drop it from textWordMethodArgs", name)
		}
	}
	for name := range textWordGameMods {
		if seen["game."+name] == 0 {
			t.Errorf("no catalog file calls game.%s — drop it from textWordGameMods", name)
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
	t.Errorf("%d colour word(s), basic land type(s) or creature type(s) hard-coded where a text-changing effect can't reach them (#3872, #3984, CR 612.2).\n"+
		"Read the word through the object whose text it is: source.TextColor(\"U\") / source.TextSubtype(\"Island\") "+
		"for a static or trigger, ctx.TextColor in an effect, or HasColorWord / HasSubtypeWord inside a predicate a "+
		"lending helper evaluates (ADR 0160's 2026-10-15 amendment). If the word really is out of reach, list the file "+
		"in textWordLiteralAllowed with the reason.\n  %s",
		len(findings), strings.Join(findings, "\n  "))
}
