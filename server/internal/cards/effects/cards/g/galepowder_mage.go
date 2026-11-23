package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Galepowder Mage — Creature — Kithkin Wizard {3}{W}, 3/3 (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"Flying
//	 Whenever this creature attacks, exile another target creature. Return that
//	 card to the battlefield under its owner's control at the beginning of the
//	 next end step."
//
// Angel of Condemnation's exile-and-return (ExileTargetsThenScheduleReturn)
// from an attack trigger; "another" is object identity, so the Mage never
// targets itself, and the target is any creature, yours or not. The delayed
// trigger returns only a card still in exile.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b2457982-6c28-47e9-9b32-9db359eb0ac5",
		Name:            "Galepowder Mage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Galepowder Mage — exile another target creature; return it at the beginning of the next end step",
					func(g *game.Game, item *game.StackItem) error {
						return effects.ExileTargetsThenScheduleReturn(effects.NewContext(g, item),
							"Galepowder Mage — return the exiled creature")
					}),
				effects.Another(effects.TargetCreature("another target creature"))),
		},
	})
}
