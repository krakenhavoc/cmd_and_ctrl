package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Demonic Tutor — "Search your library for a card, put that card
// into your hand, then shuffle."
//
// The most unconditional tutor there is, so the nil predicate is
// the card: every card in the library is a candidate and the
// controller picks one. Until S22 the primitive took the first
// library match, which made the strongest card in Magic a function
// of deck-list order.
func init() {
	Register(Spec{
		OracleID: "82004860-e589-4e38-8d61-8c0210e4ea39",
		Name:     "Demonic Tutor",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:  ctx.Controller(),
				Dest:    game.ZoneHand,
				Limit:   1,
				Reveal:  false,
				Shuffle: true,
				Reason:  "Demonic Tutor — search your library for a card",
			}.Apply(ctx)
		},
	})
}
