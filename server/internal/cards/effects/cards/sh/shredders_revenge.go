package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shredder's Revenge — {2}{B} Sorcery (Teenage Mutant Ninja Turtles, #4457):
//
//	"Choose one —
//	 • Target player discards two cards.
//	 • Target player draws two cards and loses 2 life."
//
// The first bullet is Splitting Headache's (that player chooses the cards).
// The second is Insatiable Avarice's bullet with two cards: the life loss is
// a step of its own, so it waits for a prompt the draws ask (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5dd31fcd-72db-4daf-b127-ceb27bf25a94",
		Name:         "Shredder's Revenge",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Target player discards two cards.",
				effects.TargetPlayer("target player"),
				effects.TheModesTargetDiscards(2)),
			effects.ModeWithPurpose(effects.ModeSteps("Target player draws two cards and loses 2 life.",
				effects.TargetPlayer("target player"),
				effects.TmtB6ModeTargetDraws(2),
				effects.TmtB6ModeTargetLosesLife(2)),
				effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 2, LifeLoss: 2})),
		),
	})
}
