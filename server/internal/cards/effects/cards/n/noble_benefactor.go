package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Noble Benefactor — Creature — Human Cleric {2}{U}, 2/2 (Weatherlight, S82 older sets batch 777, #5629):
//
//	"When this creature dies, each player may search their library for a card and put that card
//	 into their hand. Then each player who searched their library this way shuffles."
//
// Jungle Wayfinder's fan-out with no quality to match: the controller first and then the others
// in turn order, each search that player's own "may" for one card of any kind, put into their
// hand unrevealed, and each library shuffled by its owner's search (only a player who searched).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78c1d6fe-38e4-40c4-9252-87ff272247ad",
		Name:         "Noble Benefactor",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Noble Benefactor — each player may search their library for a card and put it into their hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					players := append([]uuid.UUID{item.Controller}, effects.OpponentsInTurnOrder(g, item.Controller)...)
					for _, p := range players {
						if err := (effects.SearchLibrary{
							Player:   p,
							Dest:     game.ZoneHand,
							Limit:    1,
							Optional: true,
							Shuffle:  true,
							Reason:   "Noble Benefactor — a card to put into your hand",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
