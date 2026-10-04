package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Conflux — Sorcery {3}{W}{U}{B}{R}{G}:
//
//	"Search your library for a white card, a blue card, a black card, a
//	 red card, and a green card. Reveal those cards, put them into your
//	 hand, then shuffle."
//
// Five searches with five predicates, run one after the other because
// one prompt cannot carry five clauses — Krosan Verge's chained shape.
// Each search leaves the library unshuffled so the next one sees the
// same order, the last one shuffles once, and failing to find a white
// card does not forfeit the blue one. Every search reveals what it
// takes, so the table sees all five.
//
// Each search is for ONE card, so a multicolored card found for white
// is already in the hand and cannot also be the blue card: the colour
// the player is asked about is the colour of the search, never a
// leftover quota. A player who wants one gold card to stand for two
// colours gets only the first, as printed ("a white card, a blue card",
// five separate cards).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "32ebb029-9f03-49c9-b9bb-cb1954e2a324",
		Name:         "Conflux",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return confluxSearchColor(ctx, 0)
		},
	})
}

// confluxColors is the five searches in printed order: the colour's
// symbol and its word for the prompt.
var confluxColors = [...]struct{ symbol, word string }{
	{"W", "white"}, {"U", "blue"}, {"B", "black"}, {"R", "red"}, {"G", "green"},
}

// confluxSearchColor runs search i, and from its continuation search
// i+1. The context is rebuilt from the live game inside each Then, the
// contract SearchLibrary.Then documents: an undo across a prompt
// resolves against the restored game.
func confluxSearchColor(ctx *Context, i int) error {
	c := confluxColors[i]
	symbol := c.symbol
	last := i == len(confluxColors)-1
	item := ctx.Item
	search := SearchLibrary{
		Player:    ctx.Controller(),
		Predicate: func(card game.Card) bool { return card.HasColor(symbol) },
		Dest:      game.ZoneHand,
		Limit:     1,
		Reveal:    true,
		Shuffle:   last,
		Reason:    "Conflux — a " + c.word + " card",
	}
	if !last {
		search.Then = func(g *game.Game, _ []uuid.UUID) error {
			return confluxSearchColor(NewContext(g, item), i+1)
		}
	}
	return search.Apply(ctx)
}
