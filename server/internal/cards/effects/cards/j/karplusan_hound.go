package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Karplusan Hound — Creature — Dog {3}{R}, 3/3 (Dominaria, S82 OLD, #5530):
//
//	"Whenever this creature attacks, if you control a Chandra planeswalker,
//	 this creature deals 2 damage to any target."
//
// An intervening "if" (CR 603.4): checked as the Hound attacks and again as
// the trigger resolves. "A Chandra planeswalker" is a planeswalker with the
// Chandra subtype, not a card merely named Chandra (Historian of Zhalfir's
// reading). The target is chosen only if the trigger goes on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7af53dd7-3994-48f1-9ae1-44380ac56ee0",
		Name:         "Karplusan Hound",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.InterveningIfYou("you control a Chandra planeswalker",
					func(g *game.Game, you uuid.UUID) bool { return effects.B42ControlsPlaneswalkerNamed(g, you, "Chandra") },
					effects.Targeting(
						effects.WheneverThisAttacks("Karplusan Hound — 2 damage to any target", effects.SourceDealsDamageToEachLegalTarget(2)),
						effects.TargetAny())),
				effects.ForTargets(effects.DamageToTarget(0, 2))),
		},
	})
}
