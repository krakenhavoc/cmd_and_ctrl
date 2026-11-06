package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Join the Dance — Sorcery {G}{W} (Avatar: The Last Airbender Eternal,
// #5075):
//
//	"Create two 1/1 white Human creature tokens.
//	 Flashback {3}{G}{W}"
//
// Army of the Damned's shape: the flashback cast from the graveyard
// keeps sorcery timing and exiles the card however it leaves the stack
// (CR 702.34a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "d8fc015c-7e6d-441d-bbeb-98ca74b21314",
		Name:             "Join the Dance",
		Purpose:          game.Purpose{Tokens: 2},
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{G}{W}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Controller: ctx.Controller(), Template: TokenCard("1/1 white Human"), N: 2}.Apply(ctx)
		},
	})
}
