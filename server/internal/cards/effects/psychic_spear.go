package effects

// Psychic Spear — Sorcery {B}:
//
//	"Target player reveals their hand. You choose a Spirit or Arcane
//	 card from it. That player discards that card."
//
// The revealed-hand pick (ADR 0116): the whole table sees the hand
// (CR 701.20a), and only a Spirit or Arcane card may be chosen (CR 701.9b).
// Spirit is a creature type, so a kindred Spirit card counts (CR
// 205.3m), and so does a changeling, which is every creature type in
// every zone (CR 702.73a). Arcane is a spell type (CR 205.3k).
// A hand with no Spirit or Arcane card is revealed and nothing is discarded
// (CR 609.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "284b0605-4b33-499e-aff3-e8b9edcce39b",
		Name:         "Psychic Spear",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve:    TargetRevealsYouChooseDiscard(Or(Subtype("Spirit"), Subtype("Arcane")), "Spirit or Arcane card"),
	})
}
