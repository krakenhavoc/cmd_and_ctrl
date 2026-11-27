package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Confound — Instant {1}{U} (Planeshift, S82 OLD #5615):
//
//	"Counter target spell that targets a creature.
//	 Draw a card."
//
// Turn Aside's clause for creatures: a spell at least one of whose targets is
// a creature on the battlefield, judged at announce and on resolution
// (CR 608.2b), then the draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ca8a0e81-f885-47ee-82bb-fb25ec61bb1c",
		Name:         "Confound",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      effects.TargetSpell("target spell that targets a creature", effects.OldB637SpellTargetsACreature()),
		OnResolve:    effects.ResolveSteps(effects.Instruction(func(ctx *effects.Context) error { return effects.SomB1CounterTheTargetSpells(ctx.Item, ctx) }), effects.DrawCards{N: 1}),
	})
}
