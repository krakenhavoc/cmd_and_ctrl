package coverage

import (
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// knownManaAbilityDisagreements are the catalogued cards where the
// printed-text classifier (game.ManaAbilityInText, CR 605.1a/b) and the
// catalog's own declarations (a ManaAbilities row or a ManaTriggers row)
// disagree, each with the reason. The classifier is the source of truth
// for "a card with a mana ability" (#3943); this list is where it was
// checked against the catalog, card by card. A row that stops applying
// fails TestManaAbilityTextAgreesWithTheCatalog until it is deleted.
var knownManaAbilityDisagreements = map[string]string{
	// The catalog is wrong: these move a card to or from a library, so
	// CR 605.1a makes each an ordinary activated ability (#3971).
	"Millikin":             "mill cost; not a mana ability by CR 605.1a (#3971)",
	"Manakin and Millikin": "mill cost; not a mana ability by CR 605.1a (#3971)",
	"Deranged Assistant":   "mill cost; not a mana ability by CR 605.1a (#3971)",
	"Charmed Pendant":      "mill cost; not a mana ability by CR 605.1a (#3971)",
	"Chromatic Sphere":     "draws a card; not a mana ability by CR 605.1a (#3971)",
	// The catalog is wrong: the card prints no mana ability (#3970).
	"Filigree Familiar": "the catalog runs a mana ability the card does not print (#3970)",
	// The classifier is right, the fixture can't show it: Dryad Arbor's
	// {T}: Add {G} is its Forest type's (CR 305.6), read off the type
	// line, which the fixture does not carry, and its reminder text is
	// not the bare "({T}: Add {G}.)" intrinsicTypeLine recognises.
	"Dryad Arbor": "a basic land type's intrinsic ability; the fixture has no type line",
}

// intrinsicReminderRE is the reminder text Scryfall prints on a land
// whose mana comes from its basic land types: "({T}: Add {U} or {B}.)",
// "({T}: Add {G}, {W}, or {U}.)".
var intrinsicReminderRE = regexp.MustCompile(`(?m)^\(\{T\}: Add ((\{[WUBRG]\})(, or |, | or )?)+\.\)$`)

// intrinsicTypeLine recovers a land's basic land types from that
// reminder text, because the fixture records text and no type line, and
// ManaAbilityInText reads CR 305.6's intrinsic ability off the type line,
// as the importer gives it. "" when the text has no such reminder.
func intrinsicTypeLine(text string) string {
	m := intrinsicReminderRE.FindString(text)
	if m == "" {
		return ""
	}
	types := map[string]string{"W": "Plains", "U": "Island", "B": "Swamp", "R": "Mountain", "G": "Forest"}
	var subtypes []string
	for _, sym := range regexp.MustCompile(`\{([WUBRG])\}`).FindAllStringSubmatch(m, -1) {
		subtypes = append(subtypes, types[sym[1]])
	}
	return "Land — " + strings.Join(subtypes, " ")
}

// TestManaAbilityTextAgreesWithTheCatalog checks the printed-text
// classifier against every catalogued face. The catalog declares a mana
// ability where it runs one (a ManaAbilities row, a ManaTriggers row) and
// nowhere else, so for a catalogued card the two should agree; where they
// don't, one of them is wrong and the row in knownManaAbilityDisagreements
// says which.
func TestManaAbilityTextAgreesWithTheCatalog(t *testing.T) {
	fx := loadOracleFixture(t)
	live := map[string]bool{}
	checked, yes := 0, 0
	for _, spec := range effects.All() {
		base, face := BaseOracleID(spec.OracleID)
		card, ok := fx[base]
		if !ok {
			continue
		}
		faces := card.AllFaces()
		if face >= len(faces) {
			continue
		}
		text := faces[face].Text
		// The engine's view: the catalog's rows, plus the intrinsic
		// ability a basic land type gives a land (CR 305.6), which the
		// engine derives from the subtypes rather than the catalog.
		typeLine := intrinsicTypeLine(text)
		catalog := len(spec.ManaAbilities) > 0 || len(spec.ManaTriggers) > 0 || typeLine != ""
		printed := game.ManaAbilityInText(typeLine, text)
		checked++
		if printed {
			yes++
		}
		if printed == catalog {
			continue
		}
		live[spec.Name] = true
		if _, ok := knownManaAbilityDisagreements[spec.Name]; ok {
			continue
		}
		t.Errorf("%s (%s): the printed text says mana ability = %v, the catalog declares %v\n\n%s\n\n"+
			"Read the card against CR 605.1a/b. Fix whichever is wrong, or pin it in\n"+
			"knownManaAbilityDisagreements with the reason.", spec.Name, spec.OracleID, printed, catalog, text)
	}
	var stale []string
	for name := range knownManaAbilityDisagreements {
		if !live[name] {
			stale = append(stale, name)
		}
	}
	sort.Strings(stale)
	for _, name := range stale {
		t.Errorf("knownManaAbilityDisagreements pins %q, which no longer disagrees; delete the row", name)
	}
	// Not vacuous: the catalog runs hundreds of mana abilities.
	if checked < 1000 || yes < 200 {
		t.Fatalf("the check compared almost nothing: %d faces, %d with a mana ability", checked, yes)
	}
	t.Logf("%d catalogued faces, %d with a printed mana ability, %d pinned", checked, yes, len(knownManaAbilityDisagreements))
}
