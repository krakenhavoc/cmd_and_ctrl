package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Early Harvest — Instant {1}{G}{G} (Planeswalker Starter Kit reprint of a
// Mirage card, S82 OLD b798, #5633):
//
//	"Target player untaps all basic lands they control."
//
// The untap is Zacama's: every tapped basic land the target player controls
// as the spell resolves, one at a time, each as its own untap. A player who
// stopped being a legal target makes the whole spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "32001382-3079-4bbe-8d86-05da654c0d1a",
		Name:         "Early Harvest",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			p, ok := ctx.ClauseTarget(0)
			if !ok || p.Kind != game.TargetPlayer {
				return nil
			}
			return effects.B16UntapAllYouControlMatching(ctx, p.ID, effects.IsBasicLand)
		},
	})
}
