package effects

// Darkwater Egg — Artifact {1} (Odyssey, #5613):
//
//	"{2}, {T}, Sacrifice this artifact: Add {U}{B}. Draw a card.
//	 (Activate only as an instant.)"
//
// One of the five Eggs, built by odyEggRow: not a mana ability, since
// it draws (CR 605.1a), so it uses the stack and an opponent may
// respond. On resolution the controller adds {U}{B}, then draws.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c4893a24-3cbc-4011-bef4-50e0c4dce16e",
		Name:         "Darkwater Egg",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{odyEggRow("{U}{B}")},
	})
}
