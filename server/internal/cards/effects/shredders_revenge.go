package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "5dd31fcd-72db-4daf-b127-ceb27bf25a94",
		Name:         "Shredder's Revenge",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Target player discards two cards.",
				TargetPlayer("target player"),
				TheModesTargetDiscards(2)),
			ModeWithPurpose(ModeSteps("Target player draws two cards and loses 2 life.",
				TargetPlayer("target player"),
				tmtB6ModeTargetDraws(2),
				tmtB6ModeTargetLosesLife(2)),
				ForTargets(game.TargetPurpose{Slot: 0, Draws: 2, LifeLoss: 2})),
		),
	})
}
