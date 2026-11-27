package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Karplusan Wolverine — Snow Creature — Wolverine Beast {R}, 1/1 (Coldsnap):
//
//	"Whenever this creature becomes blocked, you may have it deal 1 damage to
//	 any target."
//
// The target is chosen as the trigger goes on the stack and "you may" is
// asked as it resolves (CR 603.5). It triggers once each time the Wolverine
// becomes blocked, not once per blocker (CR 509.3c).
//
// No simplification.
func init() {
	const label = "Karplusan Wolverine — you may have it deal 1 damage to any target"
	effects.Register(effects.Spec{
		OracleID:     "ce690da4-6d22-4013-aef3-596d331813e4",
		Name:         "Karplusan Wolverine",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.On(game.EventBecomesBlocked, effects.Self, label,
				effects.DamageToFirstTarget(1)), effects.TargetAny()),
				"Karplusan Wolverine — have it deal 1 damage to the target?"),
		},
	})
}
