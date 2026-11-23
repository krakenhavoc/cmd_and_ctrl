package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inspiration — Instant {3}{U} (Visions, #5631):
//
//	"Target player draws two cards."
//
// Any player, the caster included. Deep Analysis's body without the
// flashback (targetPlayerDrawsTwo, target_player_draws.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f32ceb2-92c2-4dde-bf73-40bb79c3fcef",
		Name:         "Inspiration",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 2}),
		OnResolve:    effects.TargetPlayerDrawsTwo,
	})
}
