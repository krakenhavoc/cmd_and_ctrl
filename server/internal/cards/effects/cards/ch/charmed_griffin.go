package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Charmed Griffin — Creature — Griffin {3}{W}, 3/3 (Mercadian Masques,
// #5619):
//
//	"Flying
//	 When this creature enters, each other player may put an artifact or
//	 enchantment card onto the battlefield from their hand."
//
// Hunted Wumpus's enters trigger with an artifact-or-enchantment card:
// each other player in turn order is offered one from their own hand, and it
// enters under that player's control.
//
// No simplification.
func init() {
	const name = "Charmed Griffin"
	effects.Register(effects.Spec{
		OracleID:        "6a58f2a6-ff5e-4823-b3a7-49df14d64ffe",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(name+" — each other player may put an artifact or enchantment card from their hand onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, who := range effects.SeatsFromController(g, item.Controller) {
						if who == item.Controller {
							continue
						}
						if err := (effects.PutFromHandOntoBattlefield{
							Player:   who,
							Match:    effects.Or(effects.Artifact(), effects.Enchantment()),
							Optional: true,
							Label:    name + " — you may put an artifact or enchantment card from your hand onto the battlefield",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
