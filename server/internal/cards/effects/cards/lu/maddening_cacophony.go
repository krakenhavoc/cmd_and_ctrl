package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Maddening Cacophony — Sorcery {1}{U} (Secret Lair Drop):
//
//	"Kicker {3}{U}
//	 Each opponent mills eight cards. If this spell was kicked, instead each
//	 opponent mills half their library, rounded up."
//
// The kicker is the ordinary optional cost. Each opponent still in the game mills
// eight, or, if the spell was kicked, half of their own library as it is when the
// spell resolves, rounded up (CR 107.1a; Kitsune's Technique's rounding), the
// opponents in turn order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "e0f0f290-c83f-4d97-9f5d-786c7136b0d1",
		Name:          "Maddening Cacophony",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{3}{U}")},
		Purpose:       game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, id := range effects.OpponentsInTurnOrder(ctx.Game, ctx.Controller()) {
				p := ctx.PlayerByID(id)
				if p == nil || p.Eliminated || p.Library == nil {
					continue
				}
				n := 8
				if ctx.KickedTimes() > 0 {
					n = (p.Library.Size() + 1) / 2
				}
				if err := (effects.MillCards{Player: id, N: n}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
