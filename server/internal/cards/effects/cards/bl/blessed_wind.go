package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blessed Wind — Sorcery {7}{W}{W} (Prophecy, S82 OLD batch 676, #5617):
//
//	"Target player's life total becomes 20."
//
// Angel of Grace's life total reset (B31LifeBecomes) on a target player: as it
// resolves, the player gains or loses the difference (CR 119.5), so life gain
// and life loss triggers see it, and a total already 20 changes nothing. A
// player who left before resolution is an illegal target (CR 608.2b) and
// nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "93da4a35-d2d8-476c-ad03-318101c0118b",
		Name:         "Blessed Wind",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			return effects.B31LifeBecomes(ctx, target, 20)
		},
	})
}
