package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const thorncasterSliverGrant = "thorncaster-sliver/attack-ping"

// Thorncaster Sliver — Creature — Sliver {4}{R}, 2/2 (Magic 2014, #5563):
//
//	"Sliver creatures you control have "Whenever this creature attacks, it deals 1 damage to any target.""
//
// The quoted trigger is granted to each Sliver creature its controller
// controls (ADR 0093), so it is that Sliver's own ability: the Sliver that
// attacks is the one that deals the damage, and losing the grant takes the
// trigger with it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "70623348-7eda-4c68-89cd-2356b0896b79",
		Name:         "Thorncaster Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: thorncasterSliverGrant,
			Triggered: []game.TriggeredAbility{
				effects.TriggerWithPurpose(effects.Targeting(
					effects.WheneverThisAttacks("Thorncaster Sliver — it deals 1 damage to any target", effects.DamageToFirstTarget(1)),
					effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 1))),
			},
			Text: "Whenever this creature attacks, it deals 1 damage to any target.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, thorncasterSliverGrant),
		},
	})
}
