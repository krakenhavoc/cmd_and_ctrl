package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Topple — Sorcery {2}{W} (Nemesis, S82 OLD batch 679, #5618):
//
//	"Exile target creature with the greatest power among creatures on the
//	 battlefield. (If two or more creatures are tied for greatest power,
//	 target any one of them.)"
//
// The target clause is the set of creatures no creature on the battlefield
// out-powers, anyone's (effects.OldB679GreatestPowerAmongAllCreatures), read
// with the layers as they stand. It is checked again as Topple resolves (CR
// 608.2b): a creature pumped past it, or one the set no longer contains, is an
// illegal target and the spell does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75eb30cc-688d-400a-abe3-5319b79fcbae",
		Name:         "Topple",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature with the greatest power among creatures on the battlefield", effects.OldB679GreatestPowerAmongAllCreatures()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
