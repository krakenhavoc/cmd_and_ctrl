package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blisterstick Shaman — Creature — Goblin Shaman {2}{R}, 2/1 (Mirrodin Besieged,
// #5576):
//
//	"When this creature enters, it deals 1 damage to any target."
//
// Skeleton Archer's enters ping: the Shaman is the damage source, and a
// target that has become illegal by resolution is not dealt damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f76fec2-6fd0-4d33-b39a-f1684922d478",
		Name:         "Blisterstick Shaman",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Blisterstick Shaman — 1 damage to any target",
				effects.SourceDealsDamageToEachLegalTarget(1)), effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 1))),
		},
	})
}
