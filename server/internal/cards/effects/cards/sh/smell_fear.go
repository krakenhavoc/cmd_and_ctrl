package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smell Fear — Sorcery {1}{G} (Modern Horizons 2):
//
//	"Proliferate.
//	 Target creature you control fights up to one target creature you don't
//	 control."
//
// Proliferate resolves first and asks what to add; then the fight, read from
// the two target clauses. The second clause may be empty ("up to one"), in
// which case nothing fights; if either fighter is gone the fight does not
// happen (CR 701.14b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c9739382-9fe0-4d24-b58e-1e45a88491b6",
		Name:         "Smell Fear",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("up to one target creature you don't control", effects.OpponentControls()).WithCount(0, 1),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.Proliferate{}.Apply(ctx) }),
			effects.Instruction(func(ctx *effects.Context) error { return effects.SlotZeroFightsSlotOne(ctx.Game, ctx.Item) }),
		),
	})
}
