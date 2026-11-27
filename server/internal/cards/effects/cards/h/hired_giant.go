package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hired Giant — Creature — Giant {3}{R}, 4/4 (Mercadian Masques, #5619):
//
//	"When this creature enters, each other player may search their library
//	 for a land card and put that card onto the battlefield. Then each
//	 player who searched their library this way shuffles."
//
// Each other player in turn order may search their own library for any land
// card (not just a basic) and put it onto the battlefield untapped, and
// shuffles if they searched (From the Ashes' search). The Giant's controller
// searches nothing.
//
// No simplification.
func init() {
	const name = "Hired Giant"
	effects.Register(effects.Spec{
		OracleID:     "79c9a963-a428-45f1-a647-5a91ef64286e",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(name+" — each other player may search their library for a land card and put it onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, who := range effects.SeatsFromController(g, item.Controller) {
						if who == item.Controller {
							continue
						}
						if err := (effects.SearchLibrary{
							Player:    who,
							Predicate: game.Card.IsLand,
							Dest:      game.ZoneBattlefield,
							Limit:     1,
							Reveal:    true,
							Optional:  true,
							Shuffle:   true,
							Reason:    name + " — a land card, onto the battlefield",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
