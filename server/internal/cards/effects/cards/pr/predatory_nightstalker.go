package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Predatory Nightstalker — Creature — Nightstalker {3}{B}{B}, 3/2 (Portal Second Age, S82 older sets batch 734, #5625):
//
//	"When this creature enters, you may have target opponent sacrifice a creature of their choice."
//
// The opponent is the target, chosen as the trigger goes on the stack; the "may" is asked as it resolves (CR 603.5) and the opponent chooses the creature (CR 701.21).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d2ad65f-fba0-45e9-a5e6-bec63a9db979",
		Name:         "Predatory Nightstalker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(effects.WhenThisEnters("Predatory Nightstalker — you may have target opponent sacrifice a creature",
					effects.TargetPlayerSacrificesACreature("Predatory Nightstalker — sacrifice a creature")),
					"Predatory Nightstalker — have target opponent sacrifice a creature?"),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
