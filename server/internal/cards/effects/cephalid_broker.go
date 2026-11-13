package effects

// Cephalid Broker — Creature — Octopus {3}{U}, 2/2 (Odyssey, #5613):
//
//	"{T}: Target player draws two cards, then discards two cards."
//
// Cephalid Looter's ability at two, built by odyTargetPlayerLoots: the
// draws come first, and the two discards are chosen from the hand they
// made.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "18bb35a8-01ac-414f-9a45-e550060ff905",
		Name:         "Cephalid Broker",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{odyTargetPlayerLoots("{T}: Target player draws two cards, then discards two cards.", 2)},
	})
}
