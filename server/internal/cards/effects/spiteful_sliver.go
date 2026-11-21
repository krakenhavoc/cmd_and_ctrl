package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spiteful Sliver — Creature — Sliver {2}{R}, 2/2 (Commander Masters):
//
//	"Sliver creatures you control have "Whenever this creature is dealt
//	 damage, it deals that much damage to target player or planeswalker.""
//
// A granted triggered ability (ADR 0093) on your own Slivers, with Boros
// Reckoner's enrage shape: the amount is the total the Sliver was dealt in
// one batch, and the target is a player or planeswalker.
//
// No simplification.
const spitefulSliverGrant = "spiteful-sliver/reflect"

func init() {
	Register(Spec{
		OracleID:     "c4d2c804-1448-4639-b41b-0a86bac083b9",
		Name:         "Spiteful Sliver",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: spitefulSliverGrant,
			Triggered: []game.TriggeredAbility{
				TriggerWithPurpose(Targeting(cmdr2B08EnrageTrigger("Spiteful Sliver — it deals that much damage to target player or planeswalker",
					damagedCreatureReflectsChosenAmount), targetPlayerOrPlaneswalker()), ForTargets(DamageToTarget(0, 1))),
			},
			Text: "Whenever this creature is dealt damage, it deals that much damage to target player or planeswalker.",
		}},
		Static: []game.StaticAbility{
			TribalAbilityGrant(TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, spitefulSliverGrant),
		},
	})
}
