package effects

// Castigate — Sorcery {W}{B} (Guildpact, #5599):
//
//	"Target opponent reveals their hand. You choose a nonland card from
//	 it and exile that card."
//
// The revealed-hand pick with the card exiled rather than discarded
// (ChooseFromRevealedHand with Exile, ADR 0116; Soul Search without its
// rider). A hand with no nonland card is still revealed and nothing is
// exiled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "28a972a6-cf67-4ee3-aa91-f7b2549f6c48",
		Name:         "Castigate",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve:    TargetRevealsYouChooseExile(Nonland(), "nonland card"),
	})
}
