package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "c4d2c804-1448-4639-b41b-0a86bac083b9",
		Name:         "Spiteful Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: spitefulSliverGrant,
			Triggered: []game.TriggeredAbility{
				effects.TriggerWithPurpose(effects.Targeting(effects.Cmdr2B08EnrageTrigger("Spiteful Sliver — it deals that much damage to target player or planeswalker",
					effects.DamagedCreatureReflectsChosenAmount), effects.TargetPlayerOrPlaneswalker()), effects.ForTargets(effects.DamageToTarget(0, 1))),
			},
			Text: "Whenever this creature is dealt damage, it deals that much damage to target player or planeswalker.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, spitefulSliverGrant),
		},
	})
}
