package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Conqueror's Pledge — Sorcery {2}{W}{W}{W} (Zendikar, S82 OLD batch 491,
// #5583):
//
//	"Kicker {6}
//	 Create six 1/1 white Kor Soldier creature tokens. If this spell was
//	 kicked, create twelve of those tokens instead."
//
// Six tokens, or twelve when the spell was kicked (CR 702.33).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "dcc59bd6-2c5f-48eb-8008-83ce6eb5a442",
		Name:          "Conqueror's Pledge",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{6}")},
		Purpose:       game.Purpose{Tokens: 6},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := 6
			if ctx.WasKicked() {
				n = 12
			}
			return effects.CreateToken{Template: effects.TokenCard("1/1 white Kor Soldier"), N: n}.Apply(ctx)
		},
	})
}
