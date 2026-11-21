package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const regalSliverGrant = "regal-sliver/monarch"

// Regal Sliver — Creature — Sliver {3}{W}, 3/3 (Commander Masters,
// #5948):
//
//	"Sliver creatures you control have 'When this creature enters,
//	 Slivers you control get +1/+1 until end of turn if you're the
//	 monarch. Otherwise, you become the monarch.'"
//
// The monarch test is a clause of the effect, read as the trigger
// resolves (CR 608.2), so a crown lost in response still goes to the
// "otherwise" half. Each Sliver that enters carries its own trigger.
//
// No simplification.
func init() {
	t := WhenThisEnters("Regal Sliver — Slivers you control get +1/+1 until end of turn if you're the monarch, otherwise you become the monarch",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			if YoureTheMonarch(g, item.Controller) {
				return BoostUntilEOT{Match: And(Creature(), HasSubtype("Sliver"), YouControl()), Power: 1, Toughness: 1,
					Label: "Regal Sliver — Slivers you control get +1/+1"}.Apply(ctx)
			}
			return BecomeTheMonarch{}.Apply(ctx)
		})
	Register(Spec{
		OracleID:     "3a68915d-3e8b-4ede-bdd9-6c138ca63220",
		Name:         "Regal Sliver",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       regalSliverGrant,
			Triggered: []game.TriggeredAbility{t},
			Text:      "When this creature enters, Slivers you control get +1/+1 until end of turn if you're the monarch. Otherwise, you become the monarch.",
		}},
		Static: []game.StaticAbility{
			TribalAbilityGrant(TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, regalSliverGrant),
		},
	})
}
