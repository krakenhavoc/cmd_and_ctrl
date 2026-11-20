package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scour All Possibilities — Sorcery {1}{U} (Modern Horizons, #5519):
//
//	"Scry 2, then draw a card.
//	 Flashback {4}{U} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Preordain's scry-then-draw at sorcery speed: the draw is the scry's
// registered tail, so it waits for the scry answer. Flashback is Silent
// Departure's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "3fb9f0d9-6514-4005-b65a-7929f7c4df06",
		Name:             "Scour All Possibilities",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Draws: 1, Scry: 2},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{4}{U}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			c := ctx.Controller()
			return Scry{Player: c, N: 2, ThenTail: game.KeyedTail{Ref: thenDrawACard, Player: c}}.Apply(ctx)
		},
	})
}
