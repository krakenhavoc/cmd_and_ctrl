package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thirsting Roots — Sorcery {G} (Phyrexia: All Will Be One, #5479):
//
//	"Choose one —
//	 • Search your library for a basic land card, reveal it, put it into your
//	   hand, then shuffle.
//	 • Proliferate."
//
// Bushwhack's search for the first mode (failing to find is legal and the
// library is shuffled either way); the second is a plain proliferate.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c0ef21d5-f4bf-4417-b701-d98c2dbbaa43",
		Name:         "Thirsting Roots",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Search your library for a basic land card, reveal it, put it into your hand, then shuffle.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: effects.IsBasicLand,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Thirsting Roots — a basic land card",
					}.Apply(ctx)
				}),
			effects.ModeDoing("Proliferate.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.Proliferate{}.Apply(ctx)
				}),
		),
	})
}
