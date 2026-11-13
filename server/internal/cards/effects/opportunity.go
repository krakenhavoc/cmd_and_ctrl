package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Opportunity — Instant {4}{U}{U} (Urza's Legacy, #5623):
//
//	"Target player draws four cards."
//
// Path of the Enigma's draw without the vote: any player, the caster
// included, draws four. A target that has left the game is illegal and the
// spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1f544a9f-c238-4858-be19-d6cd7d023dcc",
		Name:         "Opportunity",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		Purpose:      ForTargets(game.TargetPurpose{Slot: 0, Draws: 4}),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			return DrawCards{Player: item.Targets[0].ID, N: 4}.Apply(ctx)
		},
	})
}
