package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	t := effects.WhenThisEnters("Regal Sliver — Slivers you control get +1/+1 until end of turn if you're the monarch, otherwise you become the monarch",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			if effects.YoureTheMonarch(g, item.Controller) {
				return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.HasSubtype("Sliver"), effects.YouControl()), Power: 1, Toughness: 1,
					Label: "Regal Sliver — Slivers you control get +1/+1"}.Apply(ctx)
			}
			return effects.BecomeTheMonarch{}.Apply(ctx)
		})
	effects.Register(effects.Spec{
		OracleID:     "3a68915d-3e8b-4ede-bdd9-6c138ca63220",
		Name:         "Regal Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       regalSliverGrant,
			Triggered: []game.TriggeredAbility{t},
			Text:      "When this creature enters, Slivers you control get +1/+1 until end of turn if you're the monarch. Otherwise, you become the monarch.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, regalSliverGrant),
		},
	})
}
