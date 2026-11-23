package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oreskos Explorer — Creature — Cat Scout {1}{W}, 2/2 (Commander 2015,
// #5546):
//
//	"When this creature enters, search your library for up to X Plains
//	 cards, where X is the number of players who control more lands than
//	 you. Reveal those cards, put them into your hand, then shuffle."
//
// X is counted as the trigger resolves: only an opponent can control more
// lands than you (Priest of the Blessed Graf's count). "Plains cards" is any
// card with the Plains land type, basic or not, read through the source's
// text (CR 612.2). At X = 0 you search for no cards and still shuffle
// (Celebrate the Harvest's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3055ae8-990d-495a-9cb5-971407990cfa",
		Name:         "Oreskos Explorer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Oreskos Explorer — search for up to X Plains cards", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := effects.VowB12OpponentsWithMoreLands(g, item.Controller)
				if x <= 0 {
					return g.ShuffleLibraryForEffect(item.Controller)
				}
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsLandWithSubtype("Plains"),
					Dest:      game.ZoneHand,
					Limit:     x,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Oreskos Explorer — up to X Plains cards",
				}.Apply(ctx)
			}),
		},
	})
}
