package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deadeye Quartermaster — Creature — Human Pirate {3}{U}, 2/2 (Ixalan):
//
//	"When this creature enters, you may search your library for an Equipment or
//	 Vehicle card, reveal it, put it into your hand, then shuffle."
//
// Spinewoods Armadillo's search with the "may" of Solemn Simulacrum's: the
// searcher may decline the card and the shuffle, and may find nothing (CR
// 701.23b). "Equipment" and "Vehicle" are words of the Quartermaster's text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "40abba38-533b-4f51-99e9-a2b337d65171",
		Name:         "Deadeye Quartermaster",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Deadeye Quartermaster — you may search your library for an Equipment or Vehicle card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					equipment, vehicle := ctx.TextSubtype("Equipment"), ctx.TextSubtype("Vehicle")
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.HasSubtype(equipment) || c.HasSubtype(vehicle) },
						Dest:      game.ZoneHand,
						Limit:     1,
						Optional:  true,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Deadeye Quartermaster — an Equipment or Vehicle card",
					}.Apply(ctx)
				}),
		},
	})
}
