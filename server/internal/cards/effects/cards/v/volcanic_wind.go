package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Volcanic Wind — Sorcery {4}{R}{R} (Mercadian Masques, S82 OLD, #5619):
//
//	"Volcanic Wind deals X damage divided as you choose among any number of
//	 target creatures, where X is the number of creatures on the battlefield as
//	 you cast this spell."
//
// Monstrous Onslaught's division: X is read as the spell is cast and the split is
// announced with the targets (CR 601.2d), each target getting at least 1. A
// target that is illegal on resolution takes nothing and its share is not
// reassigned. The damage is the spell's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "baa85b14-1c9f-4a19-9f07-93e3192c597d",
		Name:         "Volcanic Wind",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("any number of target creatures").WithCount(0, 0).Dividing(effects.DivideBy(effects.OldB690DivideCreaturesOnBattlefield)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DealDividedDamage(ctx)
		},
	})
}
