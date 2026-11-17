package effects

// Skycloud Egg — Artifact {1} (Odyssey, #5613):
//
//	"{2}, {T}, Sacrifice this artifact: Add {W}{U}. Draw a card.
//	 (Activate only as an instant.)"
//
// One of the five Eggs, built by odyEggRow: not a mana ability, since
// it draws (CR 605.1a), so it uses the stack and an opponent may
// respond. On resolution the controller adds {W}{U}, then draws.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1c4b6543-777e-4c3b-a9fb-5b7210d458d5",
		Name:         "Skycloud Egg",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{odyEggRow("{W}{U}")},
	})
}
