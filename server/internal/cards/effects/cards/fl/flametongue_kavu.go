package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flametongue Kavu — Creature — Kavu {3}{R}, 4/2 (Planeshift, #5615):
//
//	"When this creature enters, it deals 4 damage to target creature."
//
// Firemaw Kavu's enters trigger at 4. The target is mandatory and may be
// any creature, the Kavu included: with no other creature on the
// battlefield it must target itself (CR 603.3d), as in paper.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b215196a-3716-451f-bdbd-d16c05ace5b4",
		Name:         "Flametongue Kavu",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Flametongue Kavu — 4 damage to target creature", effects.SourceDealsDamageToEachLegalTarget(4)),
				effects.TargetCreature("target creature")), effects.ForTargets(effects.DamageToTarget(0, 4))),
		},
	})
}
