package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cerebral Confiscation — Sorcery {2}{B}:
//
//	"Choose one —
//	 • Target opponent discards two cards.
//	 • Target opponent reveals their hand. You choose a nonland card
//	   from it. That player discards that card."
//
// The first bullet is the ordinary discard: the opponent chooses
// (CR 701.9b). The second is the revealed-hand pick (ADR 0116) with
// Thoughtseize's nonland filter; a hand with no nonland card is
// revealed and nothing is discarded (CR 609.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ea71ea2-65ca-4907-b857-dbd1ba8efacd",
		Name:         "Cerebral Confiscation",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Target opponent discards two cards.",
				effects.TargetPlayer("target opponent", effects.Opponent()),
				effects.TheModesTargetDiscards(2)),
			effects.ModeDoing("Target opponent reveals their hand. You choose a nonland card from it. That player discards that card.",
				effects.TargetPlayer("target opponent", effects.Opponent()),
				effects.ModeTargetRevealsYouChooseDiscard(effects.Nonland(), "nonland card")),
		),
	})
}
