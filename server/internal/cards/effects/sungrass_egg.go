package effects

// Sungrass Egg — Artifact {1} (Odyssey, #5613):
//
//	"{2}, {T}, Sacrifice this artifact: Add {G}{W}. Draw a card.
//	 (Activate only as an instant.)"
//
// One of the five Eggs, built by odyEggRow: not a mana ability, since
// it draws (CR 605.1a), so it uses the stack and an opponent may
// respond. On resolution the controller adds {G}{W}, then draws.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "80a49a1b-a202-4c14-b093-dc76eb0f42c7",
		Name:         "Sungrass Egg",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{odyEggRow("{G}{W}")},
	})
}
