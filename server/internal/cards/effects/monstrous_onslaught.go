package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Monstrous Onslaught — Sorcery {3}{G}{G}:
//
//	"Monstrous Onslaught deals X damage divided as you choose among any
//	 number of target creatures, where X is the greatest power among
//	 creatures you control as you cast this spell."
//
// X is read as the spell is cast and the split is announced with the
// targets (CR 601.2d); each target gets at least 1. A target that is
// illegal on resolution takes nothing and its share is not reassigned.
// The damage is dealt by the spell, not by the creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "52be3da7-1e55-44c4-88bc-fb578a635ec0",
		Name:         "Monstrous Onslaught",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("any number of target creatures").WithCount(0, 0).Dividing(DivideBy(cmdrB5DivideGreatestPower)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DealDividedDamage(ctx)
		},
	})
}
