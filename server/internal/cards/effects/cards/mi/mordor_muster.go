package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mordor Muster — Sorcery {1}{B} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"You draw a card and you lose 1 life.
//	 Amass Orcs 1."
//
// Three steps of one resolution, each after any prompt the one before asked
// (CR 608.2c): the draw, the life loss, then the keyword action (CR 701.47).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "492108b7-f831-4739-925f-c3a419859c52",
		Name:         "Mordor Muster",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Amass: 1, Draws: 1, Answers: game.AnswerValue},
		OnResolve: effects.ResolveSteps(
			effects.DrawCards{N: 1},
			effects.Instruction(func(ctx *effects.Context) error {
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -1)
			}),
			effects.Amass{Subtype: "Orc", N: 1},
		),
	})
}
