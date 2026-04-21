package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Demonic Tutor — "Search your library for a card, put that card
// into your hand, then shuffle."
//
// S14 sandbox simplification: the primitive picks the FIRST library
// match. There is no "you choose" UI yet (S22 smart-cast), so the
// controller can't express a preference. In practice the library
// order matches deck-list order at start and shuffle order after
// mulligans; if a real game insists on a specific card, the player
// can physically rearrange via the deck manager UI.
func init() {
	Register(Spec{
		OracleID: "82004860-e589-4e38-8d61-8c0210e4ea39",
		Name:     "Demonic Tutor",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: func(game.Card) bool { return true },
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    false,
				Shuffle:   true,
			}.Apply(ctx)
		},
	})
}
