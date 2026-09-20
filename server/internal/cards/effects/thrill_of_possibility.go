package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thrill of Possibility — Instant for {1}{R}:
//
//	"As an additional cost to cast this spell, discard a card.
//	 Draw two cards."
//
// The reason additional costs are worth modelling rather than
// folding into OnResolve: the discard is part of casting, so it
// happens while Thrill is still on the stack. Mary Read and Anne
// Bonny sees it and makes a Treasure that's available before the
// cards are drawn; a Marauding Mako is bigger for the rest of the
// turn. Discarding on resolution would get both wrong, and would
// also let a countered Thrill keep the card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "1cb0610b-a731-42c2-b93f-0a29f63cebf4",
		Name:           "Thrill of Possibility",
		Completeness:   CompletenessFull,
		AdditionalCost: DiscardCost(1),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
		},
	})
}
