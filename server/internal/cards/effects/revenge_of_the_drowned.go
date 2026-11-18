package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Revenge of the Drowned — Instant {3}{U} (Midnight Hunt):
//
//	"Target creature's owner puts it on their choice of the top or
//	 bottom of their library. You create a 2/2 black Zombie creature
//	 token with decayed."
//
// The tuck is Lagoon Breach's (ownerTucksFirstTargetTopOrBottom): the
// owner chooses the end, and the token waits for that choice
// (ResolveSteps). The token is yours whoever owned the creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d5d869bc-8996-4d19-9e32-d1cfb968875c",
		Name:         "Revenge of the Drowned",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 1, CreatureTokens: 1},
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return ownerTucksFirstTargetTopOrBottom(ctx, "Revenge of the Drowned — put it on your choice of the top or bottom of your library")
			}),
			CreateToken{Template: TokenCard("2/2 black Zombie with decayed"), N: 1},
		),
	})
}
