package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dynamite Diver — Creature — Goblin Pilot {R}, 1/1:
//
//	"This creature saddles Mounts and crews Vehicles as though its power
//	 were 2 greater.
//	 When this creature dies, it deals 1 damage to any target."
//
// The first line is CrewsAsThoughPowerGreater (ADR 0186 Decision 4): the
// Diver counts as power 3 for crew and saddle, and its real power stays
// 1. The dies trigger is dealt by the Diver, which is in the graveyard
// by then (CR 603.10a, the same last-known source Orcish Hellraiser
// uses).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5f9ef763-22aa-4d43-b084-ef02465b489c",
		Name:         "Dynamite Diver",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			CrewsAsThoughPowerGreater(2, game.CostPowerCrew|game.CostPowerSaddle),
		},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisDies("Dynamite Diver — 1 damage to any target", sourceDealsDamageToEachLegalTarget(1)),
				TargetAny()), ForTargets(DamageToTarget(0, 1))),
		},
	})
}
