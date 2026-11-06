package coverage

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pump_x_purpose_test.go — #4976: a benefit until end of turn whose power
// or toughness is the announced X declares which side is X
// (TargetPurpose.PumpPowerIsX, PumpToughnessIsX), held to the printed
// words in both directions. The registry's own refusals are
// effects.TestPumpXGuard.

// pumpXWords finds a printed "target ... gets +A/+B ..." in one sentence.
var pumpXWords = regexp.MustCompile(`(?i)\btarget\b[^.:]*?\bgets? ([+-][0-9X]+)/([+-][0-9X]+)[^.]*`)

// xDefined is the words that make an X one the text defines rather than
// the announced one (CR 107.3a, 107.3c): "where X is the number of
// Elves", "X is that card's mana value".
var xDefined = regexp.MustCompile(`(?i)\bX is\b`)

var pumpXTargetWord = regexp.MustCompile(`(?i)\btarget\b`)

// TestPumpXDeclarationsMatchThePrintedCard holds every catalog statement
// that prints "target ... gets +X/..." or ".../+X" (a benefit; a -X/-X
// harm declares no X) to its declaration
// (#4976): when the X is the announced one, the entry for that target
// declares PumpPowerIsX for an X in the power and PumpToughnessIsX for an
// X in the toughness, and nothing for a printed number, on a benefit;
// when the sentence defines X ("where X is ..."), it declares neither,
// since the X the move names is not that number. And every declaration
// in the catalog is one this scan found. A spell's text is its oracle
// fixture (every face of it), a mode's and an activated row's its label;
// a triggered row may not declare one. Liliana of the Dark Realms's
// "+X/+X or -X/-X" declares no direction at all (#4302).
func TestPumpXDeclarationsMatchThePrintedCard(t *testing.T) {
	fx := loadOracleFixture(t)
	matched := map[string]bool{}
	announced := 0
	check := func(where, text string, clauses []game.TargetClause, p game.Purpose) {
		for _, loc := range pumpXWords.FindAllStringSubmatchIndex(text, -1) {
			words := text[loc[0]:loc[1]]
			power, toughness := text[loc[2]:loc[3]], text[loc[4]:loc[5]]
			powerX, toughnessX := power == "+X", toughness == "+X"
			if !powerX && !toughnessX || strings.HasPrefix(power, "-") || strings.HasPrefix(toughness, "-") {
				// No X, or a -X/-X harm, which declares no X (#4976
				// leaves the harm its removal price).
				continue
			}
			slot := len(pumpXTargetWord.FindAllStringIndex(text[:loc[0]], -1))
			if len(clauses) == 1 {
				slot = 0
			}
			if slot >= len(clauses) {
				continue
			}
			var e *game.TargetPurpose
			for _, x := range p.Targets.List() {
				if x.Slot == slot {
					e = &x
				}
			}
			key := fmt.Sprintf("%s slot %d", where, slot)
			if xDefined.MatchString(words) {
				if e != nil && (e.PumpPowerIsX || e.PumpToughnessIsX) {
					t.Errorf("%s declares a pump of the announced X, but %q counts X at resolution", key, words)
				}
				continue
			}
			if strings.Contains(words, " or -X") {
				continue
			}
			announced++
			matched[key] = true
			if e == nil {
				t.Errorf("%s prints %q and declares nothing for it (#4976)", key, words)
				continue
			}
			if e.PumpPowerIsX != powerX || e.PumpToughnessIsX != toughnessX {
				t.Errorf("%s prints %q and declares power X %v, toughness X %v (#4976)", key, words, e.PumpPowerIsX, e.PumpToughnessIsX)
			}
			if e.UntilEndOfTurn != game.TargetBenefit {
				t.Errorf("%s prints %q and declares %q, want a benefit", key, words, e.UntilEndOfTurn)
			}
		}
	}
	declared := func(where string, p game.Purpose) {
		for _, e := range p.Targets.List() {
			if (e.PumpPowerIsX || e.PumpToughnessIsX) && !matched[fmt.Sprintf("%s slot %d", where, e.Slot)] {
				t.Errorf("%s slot %d declares a pump of X its printed words do not give (#4976)", where, e.Slot)
			}
		}
	}
	all := effects.All()
	for _, spec := range all {
		if spec.Modes != nil {
			for i, o := range spec.Modes.Options {
				check(fmt.Sprintf("%s mode %d", spec.Name, i), o.Label, o.Targets.Clauses(), o.Purpose)
			}
		} else if spec.Targets != nil {
			base, _, _ := strings.Cut(spec.OracleID, "#")
			if card, ok := fx[base]; ok {
				for _, f := range card.AllFaces() {
					check(spec.Name, f.Text, spec.Targets.Clauses(), spec.Purpose)
				}
			}
		}
		for i, row := range spec.Activated {
			if row.Targets != nil {
				check(fmt.Sprintf("%s ability %d", spec.Name, i), row.Label, row.Targets.Clauses(), row.Purpose)
			}
		}
	}
	for _, spec := range all {
		declared(spec.Name, spec.Purpose)
		if spec.Modes != nil {
			for i, o := range spec.Modes.Options {
				declared(fmt.Sprintf("%s mode %d", spec.Name, i), o.Purpose)
			}
		}
		for i, row := range spec.Activated {
			declared(fmt.Sprintf("%s ability %d", spec.Name, i), row.Purpose)
		}
		for i, tr := range spec.Triggered {
			declared(fmt.Sprintf("%s trigger %d", spec.Name, i), tr.Purpose)
		}
	}
	t.Logf("%d statements pump a target by the announced X", announced)
	if announced < 9 {
		t.Errorf("found %d statements that pump a target by the announced X, want at least the 9 #4976 declared", announced)
	}
}
