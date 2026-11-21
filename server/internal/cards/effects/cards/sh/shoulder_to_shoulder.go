package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shoulder to Shoulder — Sorcery {2}{W} (Foundations Jumpstart):
//
//	"Support 2. (Put a +1/+1 counter on each of up to two target
//	 creatures.)
//	 Draw a card."
//
// Support as a spell (CR 701.41a: any creatures, "up to"), then the
// caster draws. With no target chosen the spell still resolves and draws.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4ffa7cc8-4774-430d-b898-5bd26366e124",
		Name:         "Shoulder to Shoulder",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.SupportTargets(2, false),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.SupportEffect(ctx.Game, item); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
