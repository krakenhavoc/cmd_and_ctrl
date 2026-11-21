package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Veteran Explorer — Creature — Human Soldier Scout {G}, 1/1 (Mystery Booster 2):
//
//	"When this creature dies, each player may search their library for up
//	 to two basic land cards, put them onto the battlefield, then shuffle."
//
// Each player searches, the controller first and then the others in turn
// order; each search is that player's own "may" (Optional) for up to two
// basic land cards, which enter UNTAPPED (the text doesn't say tapped), and
// each library is shuffled by its owner's search.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa9677d2-1535-4c57-baa8-c0408e15da47",
		Name:         "Veteran Explorer",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 2, LandsUntapped: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Veteran Explorer — each player may search for up to two basic lands",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					players := append([]uuid.UUID{item.Controller}, effects.OpponentsInTurnOrder(g, item.Controller)...)
					for _, p := range players {
						if err := (effects.SearchLibrary{
							Player:    p,
							Predicate: effects.IsBasicLand,
							Dest:      game.ZoneBattlefield,
							Limit:     2,
							Optional:  true,
							Shuffle:   true,
							Reason:    "Veteran Explorer — up to two basic lands",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
