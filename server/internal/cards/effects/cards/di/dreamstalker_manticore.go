package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dreamstalker Manticore — Enchantment Creature — Manticore {2}{R}, 4/2
// (Theros Beyond Death, #5514):
//
//	"Whenever you cast your first spell during each opponent's turn, this creature
//	 deals 1 damage to any target."
//
// The condition is B22FirstSpellOnAnOpponentsTurn (Wavebreak Hippocamp's): the
// controller's cast, on a turn that is not theirs, that the per-turn cast tally
// (bumped before EventCast) counts as their first. The target is announced as the
// trigger goes on the stack (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ecb3349-ccb8-4b54-bd81-8d649d6f96f0",
		Name:         "Dreamstalker Manticore",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WheneverYouCastFirstSpellOnAnOpponentsTurn("Dreamstalker Manticore — 1 damage to any target", effects.SourceDealsDamageToEachLegalTarget(1)),
				effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 1))),
		},
	})
}
