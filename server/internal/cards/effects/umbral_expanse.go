package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Umbral Expanse — Land — Plains Swamp (Secrets of Strixhaven Commander,
// #4834):
//
//	"({T}: Add {W} or {B}.)
//	 This land enters tapped.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The white-black member of the Amonkhet cycling-land cycle: Canyon Slough's
// shape (dualManaAbility, SelfEntersTapped, Cycling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "5a1dfc60-645d-4bc0-8883-e06ed9c80706",
		Name:          "Umbral Expanse",
		Completeness:  CompletenessFull,
		Activated:     []ActivatedAbility{Cycling("{2}")},
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("W", "B")},
	})
}
