package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c0ef21d5-f4bf-4417-b701-d98c2dbbaa43",
		Name:         "Thirsting Roots",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Search your library for a basic land card, reveal it, put it into your hand, then shuffle.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: IsBasicLand,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Thirsting Roots — a basic land card",
					}.Apply(ctx)
				}),
			ModeDoing("Proliferate.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return Proliferate{}.Apply(ctx)
				}),
		),
	})
}
