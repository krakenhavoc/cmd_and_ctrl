package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Basalt Ravager — Creature — Giant Wizard {3}{R}, 4/2 (Kaldheim):
//
//	"When this creature enters, it deals X damage to any target, where X is
//	 the greatest number of creatures you control that have a creature type in
//	 common."
//
// X is counted as the trigger resolves (S82OldB064GreatestCreaturesSharingAType):
// the Ravager itself counts, and a changeling counts for every type. The
// target is chosen as the trigger goes on the stack; if it is gone nothing
// happens (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "980c32ff-c154-4ac9-adc2-ed3a85c8d71c",
		Name:         "Basalt Ravager",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Basalt Ravager — it deals X damage to any target, where X is the greatest number of creatures you control that have a creature type in common",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					ts := ctx.LegalTargets()
					if len(ts) == 0 {
						return nil
					}
					x := effects.S82OldB064GreatestCreaturesSharingAType(g, item.Controller)
					return effects.DealDamage{Source: item.SourceCardID, Target: ts[0].ID, Amount: x}.Apply(ctx)
				}),
				effects.TargetAny()),
		},
	})
}
