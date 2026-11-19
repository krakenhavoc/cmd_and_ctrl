package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Corpsehatch — Sorcery {3}{B}{B} (Rise of the Eldrazi, #5581):
//
//	"Destroy target nonblack creature. Create two 0/1 colorless Eldrazi
//	 Spawn creature tokens. They have "Sacrifice this token: Add {C}.""
//
// The tokens are made whether or not the creature is destroyed
// (indestructible, regenerated); with its only target illegal the spell
// doesn't resolve at all (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5a6c9376-313b-4420-890e-5aa78fa6382f",
		Name:         "Corpsehatch",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target nonblack creature", NonBlack()),
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 2, CreatureTokens: 2},
		OnResolve: ResolveSteps(Instruction(func(ctx *Context) error {
			return destroyFirstLegalCardTarget(ctx.Game, ctx.Item)
		}), roeB1Spawns(2)),
	})
}
