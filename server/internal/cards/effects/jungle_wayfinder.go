package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jungle Wayfinder — Creature — Elf Warrior {2}{G}, 3/3 (Battlebond,
// #5529):
//
//	"When this creature enters, each player may search their library for
//	 a basic land card, reveal it, put it into their hand, then shuffle."
//
// Veteran Explorer's shape, to the hand: each player searches, the
// controller first and then the others in turn order, each search that
// player's own "may" (Optional) for one basic land card, revealed and put
// into their hand, and each library shuffled by its owner's search.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "794ad296-4718-45b7-ab85-6ae45903ada7",
		Name:         "Jungle Wayfinder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Jungle Wayfinder — each player may search for a basic land card and put it into their hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					players := append([]uuid.UUID{item.Controller}, opponentsInTurnOrder(g, item.Controller)...)
					for _, p := range players {
						if err := (SearchLibrary{
							Player:    p,
							Predicate: IsBasicLand,
							Dest:      game.ZoneHand,
							Limit:     1,
							Reveal:    true,
							Optional:  true,
							Shuffle:   true,
							Reason:    "Jungle Wayfinder — a basic land card",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
