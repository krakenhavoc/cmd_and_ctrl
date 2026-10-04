package effects

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
	Register(Spec{
		OracleID:     "1ea71ea2-65ca-4907-b857-dbd1ba8efacd",
		Name:         "Cerebral Confiscation",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Target opponent discards two cards.",
				TargetPlayer("target opponent", Opponent()),
				TheModesTargetDiscards(2)),
			ModeDoing("Target opponent reveals their hand. You choose a nonland card from it. That player discards that card.",
				TargetPlayer("target opponent", Opponent()),
				ModeTargetRevealsYouChooseDiscard(Nonland(), "nonland card")),
		),
	})
}
