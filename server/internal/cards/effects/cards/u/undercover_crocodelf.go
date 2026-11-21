package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Undercover Crocodelf — Creature — Elf Crocodile Detective {4}{G}{U}
// (Murders at Karlov Manor, #4469):
//
//	"Whenever this creature deals combat damage to a player, investigate.
//	 Disguise {3}{G/U}{G/U} (You may cast this card face down for {3} as a
//	 2/2 creature with ward {2}. Turn it face up any time for its disguise
//	 cost.)"
//
// Investigate is a Clue token (CR 701.16a). A face-down Crocodelf has no
// text and so no trigger (CR 708.2a): the ability is there once it is face up.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "3ba223b7-743f-4f99-836e-869778fb2e99",
		Name:             "Undercover Crocodelf",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Disguise("{3}{G/U}{G/U}")},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Undercover Crocodelf — investigate",
				effects.Do(effects.Investigate{N: 1})),
		},
	})
}
