package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
			controller := item.Controller
			return SearchLibrary{
				Player: controller,
				// nil predicate is "a card" — every card in the
				// library is a candidate, which is what makes this
				// the cheapest unconditional tutor in the game.
				Dest:    game.ZoneHand,
				Limit:   1,
				Shuffle: true,
				Reason:  "Gamble — search your library for a card",
				Then: func(g *game.Game, _ []uuid.UUID) error {
					// "THEN discard a card at random." The discard has
					// to be chained: since S22 the search returns while
					// its prompt is open, and discarding before the
					// tutored card is in hand would take the random
					// card from the wrong hand — which is precisely
					// the joke the card is built on.
					return g.DiscardRandomForEffect(controller, 1)
				},
			}.Apply(ctx)
		},
	})
}
