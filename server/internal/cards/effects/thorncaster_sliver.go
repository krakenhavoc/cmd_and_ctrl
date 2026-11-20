package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "70623348-7eda-4c68-89cd-2356b0896b79",
		Name:         "Thorncaster Sliver",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: thorncasterSliverGrant,
			Triggered: []game.TriggeredAbility{
				TriggerWithPurpose(Targeting(
					WheneverThisAttacks("Thorncaster Sliver — it deals 1 damage to any target", DamageToFirstTarget(1)),
					TargetAny()), ForTargets(DamageToTarget(0, 1))),
			},
			Text: "Whenever this creature attacks, it deals 1 damage to any target.",
		}},
		Static: []game.StaticAbility{
			TribalAbilityGrant(TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, thorncasterSliverGrant),
		},
	})
}
