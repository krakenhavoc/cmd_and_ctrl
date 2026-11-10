package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unexpected Assistance — Instant {3}{U}{U} (Lorwyn Eclipsed, #4458):
//
//	"Convoke
//	 Draw three cards, then discard a card."
//
// Convoke is the shared tap cost (CR 702.51), applied after the total cost
// is determined. The body is the loot shape: the three draws land first,
// so a drawn card is a legal discard, and the discard is the controller's
// choice (drawThenDiscard).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c452e748-8a2a-4f68-8869-e294925a8b0b",
		Name:         "Unexpected Assistance",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return drawThenDiscard(ctx.Game, item.Controller, item.SourceCardID, 3, 1, "")
		},
	})
}
