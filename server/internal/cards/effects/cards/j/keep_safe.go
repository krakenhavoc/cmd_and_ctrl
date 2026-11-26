package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keep Safe — Instant {1}{U} (Ikoria: Lair of Behemoths, #5512):
//
//	"Counter target spell that targets a permanent you control.
//	 Draw a card."
//
// The target is a spell on the stack with a target that is a permanent you
// control (Turn Aside's clause, for any permanent, whoever cast the spell).
// If the target spell is gone as Keep Safe resolves, the whole spell does
// nothing and you draw no card (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5602c14d-cc22-47d7-a1b0-6491749f00dd",
		Name:         "Keep Safe",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetSpell("target spell that targets a permanent you control", effects.B14SpellTargetsAPermanentYouControl()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.SomB1CounterTheTargetSpells(ctx.Item, ctx) }),
			effects.DrawCards{N: 1},
		),
	})
}
