package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unexplained Vision — Sorcery {4}{U} (Throne of Eldraine, #5516):
//
//	"Draw three cards.
//	 Adamant — If at least three blue mana was spent to cast this spell, scry 3."
//
// The draw, then the adamant scry as a second step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "9f693483-f4df-497a-a6c7-fc269fc60ea8",
		Name:           "Unexplained Vision",
		Completeness:   CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"U"},
		Purpose:        game.Purpose{Draws: 3},
		OnResolve: ResolveSteps(
			DrawCards{N: 3},
			eldB1IfAdamant("U", Scry{N: 3}),
		),
	})
}
