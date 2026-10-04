package effects

// Auntie's Sentence — Sorcery {1}{B}:
//
//	"Choose one —
//	 • Target opponent reveals their hand. You choose a nonland
//	   permanent card from it. That player discards that card.
//	 • Target creature gets -2/-2 until end of turn."
//
// The first bullet is the revealed-hand pick (ADR 0116) filtered to
// nonland permanent cards: an artifact, creature, enchantment,
// planeswalker or battle card, read as it is in the hand (a modal
// double-faced card by its front face, CR 712.8a). A hand with none is
// revealed and nothing is discarded (CR 609.3). The second is an
// ordinary pump on the bullet's own target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "406e6db7-6e31-44b9-a793-57ee10688f71",
		Name:         "Auntie's Sentence",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Target opponent reveals their hand. You choose a nonland permanent card from it. That player discards that card.",
				TargetPlayer("target opponent", Opponent()),
				ModeTargetRevealsYouChooseDiscard(NonlandPermanentCard(), "nonland permanent card")),
			ModeDoing("Target creature gets -2/-2 until end of turn.",
				TargetCreature("target creature"),
				BoostTheModesTarget(-2, -2, "Auntie's Sentence")),
		),
	})
}
