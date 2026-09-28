package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Increasing Devotion — Sorcery {3}{W}{W}:
//
//	"Create five 1/1 white Human creature tokens. If this spell was
//	 cast from a graveyard, create ten of those tokens instead.
//	 Flashback {7}{W}{W} (You may cast this card from your graveyard
//	 for its flashback cost. Then exile it.)"
//
// The one flashback card in this batch whose resolution reads back
// WHICH cost paid for it: ctx.PaidAltCost("flashback") is exactly "if
// this spell was cast from a graveyard" — this card has only the one
// alternative cost and it is bound to the graveyard (FromZone), so
// "paid flashback" and "cast from a graveyard" are the same fact here.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "7a5ff4d4-27b7-47d4-ba88-970c63c4e3fb",
		Name:             "Increasing Devotion",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{7}{W}{W}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := 5
			if ctx.PaidAltCost("flashback") {
				n = 10
			}
			return CreateToken{Controller: ctx.Controller(), Template: TokenCard("1/1 white Human"), N: n}.Apply(ctx)
		},
	})
}
