package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unexpected Windfall — Instant for {2}{R}{R}:
//
//	"As an additional cost to cast this spell, discard a card.
//	 Draw two cards and create two Treasure tokens."
//
// Functionally identical to Big Score at a different mana cost —
// the deck runs both because the effect is the ritual half of its
// draw engine, not because they differ.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "498c10c9-253d-4b15-b48c-1509381b17e8",
		Name:           "Unexpected Windfall",
		Completeness:   CompletenessFull,
		AdditionalCost: DiscardCost(1),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return drawAndTreasure(item, ctx, 2, 2)
		},
	})
}

// drawAndTreasure is the shared body of Big Score and Unexpected
// Windfall: draw `cards`, then create `treasures` Treasure tokens,
// in printed order.
func drawAndTreasure(item *game.StackItem, ctx *Context, cards, treasures int) error {
	if err := (DrawCards{Player: item.Controller, N: cards}).Apply(ctx); err != nil {
		return err
	}
	return CreateToken{
		Controller: item.Controller,
		Template:   TreasureToken(),
		N:          treasures,
	}.Apply(ctx)
}
