package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Big Score — Instant for {3}{R}:
//
//	"As an additional cost to cast this spell, discard a card.
//	 Draw two cards and create two Treasure tokens."
//
// Thrill of Possibility with a rider. Four permanents' worth of
// artifact triggers in a deck that cares (Reckless Fireweaver,
// Ingenious Artillerist) — two Treasures entering, plus whatever
// the discard itself sets off.
func init() {
	Register(Spec{
		OracleID:       "a5cbd257-c836-493e-bb1a-76242619dea2",
		Name:           "Big Score",
		AdditionalCost: DiscardCost(1),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return drawAndTreasure(item, ctx, 2, 2)
		},
	})
}
