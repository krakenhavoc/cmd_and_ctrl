package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:      "5a1dfc60-645d-4bc0-8883-e06ed9c80706",
		Name:          "Umbral Expanse",
		Completeness:  effects.CompletenessFull,
		Activated:     []effects.ActivatedAbility{effects.Cycling("{2}")},
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("W", "B")},
	})
}
