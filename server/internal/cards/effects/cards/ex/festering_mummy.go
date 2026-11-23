package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Festering Mummy — Creature — Zombie {B}, 1/1
// (Amonkhet, #5535):
//
//	"When this creature dies, you may put a -1/-1 counter on target creature."
//
// Necropede's dies trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bdf78365-5887-497e-af25-ca0cb731f35a",
		Name:         "Festering Mummy",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisDies("Festering Mummy — you may put a -1/-1 counter on target creature",
					effects.PutACounterOnTheTarget(game.CounterMinusOne)),
				effects.Harmful(effects.TargetCreature("target creature"))),
				"Festering Mummy — put a -1/-1 counter on the target creature?"),
		},
	})
}
