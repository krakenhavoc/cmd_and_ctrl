package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Savage Conception — Sorcery {3}{G}{G}:
//
//	"Create a 3/3 green Beast creature token.
//	 Retrace"
//
// Retrace (CR 702.81) is a graveyard cast for the PRINTED mana cost plus
// a land card discarded from hand: `CastableZones` opens the graveyard and
// `Retrace(cost)` prices it (ADR 0066, 2026-10-07 amendment). The spell goes
// back to the graveyard when it resolves, so it can be retraced again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "494050f4-0a55-415d-9ae9-feb17e61c4e1",
		Name:             "Savage Conception",
		Purpose:          game.Purpose{Tokens: 1},
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Retrace("{3}{G}{G}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{
				Controller: ctx.Controller(),
				Template:   effects.TokenCard("3/3 green Beast"),
				N:          1,
			}.Apply(ctx)
		},
	})
}
