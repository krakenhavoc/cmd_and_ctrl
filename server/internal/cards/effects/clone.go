package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Clone — "You may have this creature enter as a copy of any
// creature on the battlefield."
//
// The baseline of the copy-effect class: no "except" clause, no
// controller restriction, no type restriction beyond "creature". If
// the player declines — or there is no creature to copy — Clone
// enters as its printed 0/0.
//
// And then it dies (CR 704.5f), as the printed card does. It used to
// survive, and said so in a caveat: the toughness state-based action
// skipped every creature with printed toughness 0 and no counters,
// because the same shape is what the importer writes for stats it
// cannot parse. #691 narrowed the skip to objects with no printing
// behind them (game.Card.ToughnessIsKnown), and a Clone has one.
//
// "Any creature on the battlefield" is not targeting: a copy choice
// is made as the permanent enters, so hexproof, shroud and
// protection do not stop it (CR 707.2 — the ruling that makes Clone
// the answer to an opposing Blightsteel Colossus).
func init() {
	Register(Spec{
		OracleID:     "42226b87-0746-4ebf-9fd0-108d508462af",
		Name:         "Clone",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf(
				"Clone",
				anyCreatureOnBattlefield,
				nil,
			),
		},
	})
}
