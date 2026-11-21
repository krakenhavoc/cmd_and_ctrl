package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b39_helpers.go — shared pieces for the Assassin's Creed cards of S82
// GAP batch 39 (#5948). Append-only.

// gapB39ThopterCopyException is "except it's a 0/2 Thopter artifact creature
// with flying in addition to its other types" on a token copy's template
// (CR 707.9b): base power and toughness 0/2 (printed values, not a `*`
// stand-in), the artifact and creature types and the Thopter subtype added to
// the copied ones, and flying added to its abilities.
func GapB39ThopterCopyException(thopter string) func(tok *game.Card) {
	return func(tok *game.Card) {
		tok.Power, tok.Toughness = 0, 2
		tok.VariableToughness = false
		tok.PrintedPTKnown = true
		super, types, subs := game.ParseTypeLine(tok.TypeLine)
		have := map[string]bool{}
		for _, t := range types {
			have[strings.ToLower(t)] = true
		}
		head := append(append([]string(nil), super...), types...)
		for _, add := range []string{"Artifact", "Creature"} {
			if !have[strings.ToLower(add)] {
				head = append(head, add)
			}
		}
		hasSub := false
		for _, s := range subs {
			if strings.EqualFold(s, thopter) {
				hasSub = true
			}
		}
		if !hasSub {
			subs = append(append([]string(nil), subs...), thopter)
		}
		tok.TypeLine = strings.Join(head, " ") + " — " + strings.Join(subs, " ")
		TokenCopyGainsKeyword(tok, "flying")
	}
}

// TokenCopyGainsKeyword is "except it has <keyword>" on a token copy's template
// (CR 707.9a): the keyword joins the template's printed Keywords once.
func TokenCopyGainsKeyword(t *game.Card, keyword string) {
	for _, kw := range t.Keywords {
		if kw == keyword {
			return
		}
	}
	t.Keywords = append(t.Keywords, keyword)
}
