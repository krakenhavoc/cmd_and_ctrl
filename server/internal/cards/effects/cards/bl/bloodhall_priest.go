package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodhall Priest — Creature — Vampire Cleric {2}{B}{R}, 4/4 (Innistrad
// Remastered, S82 GAP, #5948):
//
//	"Whenever this creature enters or attacks, if you have no cards in
//	 hand, this creature deals 2 damage to any target.
//	 Madness {1}{B}{R}"
//
// One printed ability that watches two events: an enters-or-attacks
// trigger (Phlage's shape) behind an intervening "if" (CR 603.4) on the
// controller's empty hand, checked as it triggers and again as it
// resolves. The target is chosen as the trigger goes on the stack. The
// damage comes from the Priest, and from its last-known information if
// it has left by resolution (CR 608.2h). Madness is the shared keyword
// (Spec.Madness).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2eed5eb1-ab67-4f23-885d-922131364aba",
		Name:         "Bloodhall Priest",
		Completeness: effects.CompletenessFull,
		Madness:      "{1}{B}{R}",
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.InterveningIf(effects.CondGapB09YouHaveNoCardsInHand{},
				effects.Targeting(effects.WhenThisEntersOrAttacks("Bloodhall Priest — if you have no cards in hand, 2 damage to any target",
					effects.DamageToFirstTarget(2)),
					effects.TargetAny())),
				effects.ForTargets(effects.DamageToTarget(0, 2))),
		},
	})
}
