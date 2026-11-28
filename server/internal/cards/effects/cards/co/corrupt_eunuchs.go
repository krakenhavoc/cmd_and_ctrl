package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Corrupt Eunuchs — Creature — Human Advisor {3}{R}, 2/2 (Portal Three Kingdoms,
// S82 #5622):
//
//	"When this creature enters, it deals 2 damage to target creature."
//
// Akoum Boulderfoot's trigger over a creature target: the damage is dealt by the
// Eunuchs, still as it last existed there if it has left before the trigger
// resolves, and a target that is illegal by then takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77e604fe-d6ca-495f-ba22-a7d8ba88c8b5",
		Name:         "Corrupt Eunuchs",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Corrupt Eunuchs — 2 damage to target creature",
				effects.SourceDealsDamageToEachLegalTarget(2)), effects.TargetCreature("target creature")), effects.ForTargets(effects.DamageToTarget(0, 2))),
		},
	})
}
