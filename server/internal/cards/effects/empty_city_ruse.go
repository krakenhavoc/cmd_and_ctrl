package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Empty City Ruse — {W} Sorcery (Avatar: The Last Airbender Eternal,
// #5075):
//
//	"Target opponent skips all combat phases of their next turn."
//
// TargetPlayerSkipsCombatPhasesOfNextTurn (skip_combat_phases.go): the
// skip waits for the opponent's next turn that really begins, extra or
// normal, and then no combat phase of that turn happens, one added later
// in it included (CR 500.11, 614.10). A second Ruse on the same player
// skips the combat phases of their turn after that (CR 614.10a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b0391ea4-e37d-48c9-91df-79aefe376bcc",
		Name:         "Empty City Ruse",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return TargetPlayerSkipsCombatPhasesOfNextTurn{}.Apply(ctx)
		},
	})
}
