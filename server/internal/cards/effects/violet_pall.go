package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Violet Pall — Kindred Instant — Faerie {4}{B} (Morningtide, #5592):
//
//	"Destroy target nonblack creature. Create a 1/1 black Faerie Rogue
//	 creature token with flying."
//
// Corpsehatch's shape: the destroy, then the token. With its only target
// gone the spell does nothing at all (CR 608.2b), so no token either.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bf6ea336-6a4d-4653-acdd-f09af0b9f0e3",
		Name:         "Violet Pall",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target nonblack creature", NonBlack()),
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 1, CreatureTokens: 1},
		OnResolve: ResolveSteps(Instruction(func(ctx *Context) error {
			return destroyFirstLegalCardTarget(ctx.Game, ctx.Item)
		}), CreateToken{Template: FaerieRogueToken(), N: 1}),
	})
}
