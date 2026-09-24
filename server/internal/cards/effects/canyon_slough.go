package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Canyon Slough — Land — Swamp Mountain (EDHREC rank 4457):
//
//	"({T}: Add {B} or {R}.)
//	 This land enters tapped.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The black-red member of the Amonkhet cycling-land cycle — same
// shape as Polluted Mire, with the mana ability produced two colours
// instead of one (dualManaAbility, the pipe-syntax two-colour
// declaration the surveil lands and the verges already use).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "2031b17c-0536-446f-a9aa-b46fe79b7ea7",
		Name:          "Canyon Slough",
		Completeness:  CompletenessFull,
		Activated:     []ActivatedAbility{Cycling("{2}")},
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("B", "R")},
	})
}
