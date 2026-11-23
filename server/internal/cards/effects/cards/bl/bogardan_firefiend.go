package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bogardan Firefiend — Creature — Elemental Spirit {2}{R}, 2/1 (Weatherlight, #5629):
//
//	"When this creature dies, it deals 2 damage to target creature."
//
// Munitions Enthusiast's dies trigger with a creature target: the
// Firefiend deals the damage as it last existed (CR 608.2h), and the
// target is chosen as the trigger goes on the stack (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1d61f3a7-5861-479b-8312-f2d308c35b68",
		Name:         "Bogardan Firefiend",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisDies("Bogardan Firefiend — 2 damage to target creature", effects.SourceDealsDamageToEachLegalTarget(2)),
				effects.TargetCreature("target creature")), effects.ForTargets(effects.DamageToTarget(0, 2))),
		},
	})
}
