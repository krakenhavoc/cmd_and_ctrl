package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reckless Handling — Sorcery {1}{R} (EDHREC rank 2705):
//
//	"Search your library for an artifact card, reveal it, put it into
//	 your hand, shuffle, then discard a card at random. If an artifact
//	 card was discarded this way, Reckless Handling deals 2 damage to
//	 each opponent."
//
// Gamble for artifacts, with a consolation prize when the random
// discard eats the tutor target. The search is the S22 chooser
// (reveal, to hand, shuffle), and everything after "then" is chained
// in the search's continuation — the random discard has to see the
// tutored card in hand, which is the joke the card is built on. The
// discarded card is read back off the discard event the random
// discard just logged (b25DiscardedByAfter, bounded to events after
// the search so an older discard is never mistaken for this one),
// and an artifact — the tutored card or any other artifact in hand —
// deals 2 to each opponent. An empty hand discards nothing and deals
// nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f1ea7dc5-01cb-4780-947c-08ea9324c52f",
		Name:         "Reckless Handling",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := item.Controller
			return SearchLibrary{
				Player:    controller,
				Predicate: func(c game.Card) bool { return c.IsArtifact() },
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Reckless Handling — search your library for an artifact card",
				Then: func(g *game.Game, _ []uuid.UUID) error {
					mark := b25LastEventSeq(g)
					if err := g.DiscardRandomForEffect(controller, 1); err != nil {
						return err
					}
					discarded, ok := b25DiscardedByAfter(g, controller, mark)
					if !ok || !discarded.IsArtifact() {
						return nil
					}
					return damageToEachOpponent(g, item, 2)
				},
			}.Apply(ctx)
		},
	})
}
