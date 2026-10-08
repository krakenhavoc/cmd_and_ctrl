package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cenn's Enlistment — Sorcery {3}{W}:
//
//	"Create two 1/1 white Kithkin Soldier creature tokens.
//	 Retrace"
//
// Retrace (CR 702.81) is a graveyard cast for the PRINTED mana cost plus
// a land card discarded from hand: see Spitting Image and ADR 0066's
// 2026-10-07 amendment. The spell goes back to the graveyard when it
// resolves, so it can be retraced again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "49fe9f5a-5821-4586-b913-7d8aef1f8669",
		Name:             "Cenn's Enlistment",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Retrace("{3}{W}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: TokenCard("1/1 white Kithkin Soldier"), N: 2}.Apply(ctx)
		},
	})
}
