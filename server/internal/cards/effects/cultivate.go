package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cultivate — "Search your library for up to two basic land cards,
// reveal those cards, put one onto the battlefield tapped and the
// other into your hand, then shuffle."
//
// S14 sandbox simplifications:
//   - Both lands enter UNTAPPED. Enters-tapped is a replacement-
//     effect concern that lands alongside the layer / replacement
//     pipeline in S17. Matches the posture Path to Exile's land-
//     fetch adopts.
//   - Auto-picks the first two basic lands in library order. "Up to
//     two" degrades gracefully: a library with one basic puts that
//     one onto the battlefield and skips the hand grab; a library
//     with zero basics no-ops both searches. The SearchLibrary
//     primitive already returns nil on predicate-miss.
//   - Two sequential SearchLibrary calls: the first with shuffle
//     deferred (so the second sees the remaining basics in order),
//     the second with shuffle true (the rules require exactly one
//     shuffle at the end).
func init() {
	Register(Spec{
		OracleID: "8b755881-a72d-4e21-a369-d2924eb4585a",
		Name:     "Cultivate",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			if err := (SearchLibrary{
				Player:    controller,
				Predicate: IsBasicLand,
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Reveal:    true,
				Shuffle:   false,
			}).Apply(ctx); err != nil {
				return err
			}
			return SearchLibrary{
				Player:    controller,
				Predicate: IsBasicLand,
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
			}.Apply(ctx)
		},
	})
}
