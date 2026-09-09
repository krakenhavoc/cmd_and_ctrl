package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gamble — Sorcery for {R}:
//
//	"Search your library for a card, put that card into your hand,
//	 discard a card at random, then shuffle."
//
// The cheapest unconditional tutor in the game, and the drawback is
// the joke: with one card in hand after the search, the random
// discard is the card you just found. This deck doesn't mind — a
// discarded Pirate is a Treasure off the commander, and the Mako
// grows either way.
//
// The random discard is genuinely random here rather than a prompt,
// which is the one place the sandbox's RNG is load-bearing for
// fairness.
func init() {
	Register(Spec{
		OracleID: "a54f0869-94c8-42af-9080-166efb9486a4",
		Name:     "Gamble",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (SearchLibrary{
				Player:    item.Controller,
				Predicate: func(game.Card) bool { return true },
				Dest:      game.ZoneHand,
				Limit:     1,
				Shuffle:   true,
			}).Apply(ctx); err != nil {
				return err
			}
			return DiscardCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
