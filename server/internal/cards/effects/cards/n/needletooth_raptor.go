package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Needletooth Raptor — Creature — Dinosaur {3}{R}, 2/2 (Rivals of Ixalan,
// #5531):
//
//	"Enrage — Whenever this creature is dealt damage, it deals 5 damage to
//	 target creature an opponent controls."
//
// The shared enrage trigger with a target creature an opponent controls,
// picked as the trigger goes on the stack (CR 603.3d) and re-checked as it
// resolves (CR 608.2b). The Raptor is the source, even when the damage
// killed it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c431df9-095a-44f1-993e-74eb215d75fe",
		Name:         "Needletooth Raptor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Cmdr2B08EnrageTrigger("Needletooth Raptor — Enrage: 5 damage to target creature an opponent controls",
				effects.DealDamageToTheTarget(5)), effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
