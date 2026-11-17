package effects

// Mossfire Egg — Artifact {1} (Odyssey, #5613):
//
//	"{2}, {T}, Sacrifice this artifact: Add {R}{G}. Draw a card.
//	 (Activate only as an instant.)"
//
// One of the five Eggs, built by odyEggRow: not a mana ability, since
// it draws (CR 605.1a), so it uses the stack and an opponent may
// respond. On resolution the controller adds {R}{G}, then draws.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "364b3231-c0e7-45a8-90b9-a1cdb584dd7c",
		Name:         "Mossfire Egg",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{odyEggRow("{R}{G}")},
	})
}
