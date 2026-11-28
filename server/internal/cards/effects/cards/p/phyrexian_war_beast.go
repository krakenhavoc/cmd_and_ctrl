package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phyrexian War Beast — Artifact Creature — Phyrexian Beast {3} 3/4
// (Alliances, S82 OLD batch 819):
//
//	"When this creature leaves the battlefield, sacrifice a land and this
//	 creature deals 1 damage to you."
//
// "You" is the controller it had as it left (last-known information, CR
// 608.2h), and the damage is dealt by the creature that left.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c93c418-e10b-4a6d-bddb-3b786bb6a928",
		Name:         "Phyrexian War Beast",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisLeaves("Phyrexian War Beast — sacrifice a land and it deals 1 damage to you",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					g.PlayerSacrificesForEffect(item.SourceCardID, item.Controller, effects.SacrificeSpec("a land", effects.Land()),
						"Phyrexian War Beast — sacrifice a land")
					return effects.DealDamage{Source: ctx.Source(), Target: item.Controller, Amount: 1}.Apply(ctx)
				}),
		},
	})
}
