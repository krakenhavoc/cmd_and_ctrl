package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Subterranean Scout — Creature — Goblin Scout {1}{R}, 2/1 (Magic Origins, #5549):
//
//	"When this creature enters, target creature with power 2 or less can't be blocked this turn."
//
// Spider-Man, Hometown Hero's entry: the power bound is checked as the
// trigger goes on the stack and again as it resolves; the restriction
// ends in the cleanup step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "19f66bfa-bdfc-430a-ae39-4535f0b28324",
		Name:         "Subterranean Scout",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.Targeting(
					effects.WhenThisEnters("Subterranean Scout — target creature with power 2 or less can't be blocked this turn",
						effects.NemTargetRestrictedThisTurn(game.CantBeBlocked, "Subterranean Scout — can't be blocked")),
					effects.TargetCreature("target creature with power 2 or less", effects.PowerLE(2))),
				effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0))),
		},
	})
}
