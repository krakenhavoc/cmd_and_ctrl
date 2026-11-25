package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Humble — Instant {1}{W} (Eternal Masters, S82 OLD, #5542):
//
//	"Until end of turn, target creature loses all abilities and has base power
//	 and toughness 0/1."
//
// Ovinize's single continuous effect: layer 6 removes the abilities, layer 7b
// sets the base power and toughness (CR 613.1), so counters and later pumps
// still add. A target that has left makes it do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2d41f12-18d8-473b-9308-3f02d7d791f0",
		Name:         "Humble",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerRestrict},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.LosesAbilitiesAndBasePTUntilEOT(ctx, "Humble", 0, 1)
		},
	})
}
