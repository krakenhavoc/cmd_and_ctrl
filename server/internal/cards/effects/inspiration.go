package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inspiration — Instant {3}{U} (Visions, #5631):
//
//	"Target player draws two cards."
//
// Any player, the caster included. Deep Analysis's body without the
// flashback (targetPlayerDrawsTwo, target_player_draws.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8f32ceb2-92c2-4dde-bf73-40bb79c3fcef",
		Name:         "Inspiration",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		Purpose:      ForTargets(game.TargetPurpose{Slot: 0, Draws: 2}),
		OnResolve:    targetPlayerDrawsTwo,
	})
}
