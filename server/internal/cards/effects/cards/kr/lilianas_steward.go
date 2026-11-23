package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Liliana's Steward — Creature — Zombie {B}, 1/2 (Jumpstart, #4892):
//
//	"{T}, Sacrifice this creature: Target opponent discards a card.
//	 Activate only as a sorcery."
//
// The opponent chooses the card they discard (CR 701.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc1203cb-6758-4454-b78b-341930ff4da1",
		Name:         "Liliana's Steward",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{T}, Sacrifice this creature: Target opponent discards a card. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			SorcerySpeed: true,
			Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect:       effects.TargetOpponentDiscards(1),
		}},
	})
}
