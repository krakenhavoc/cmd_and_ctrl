package effects

// Shadowblood Egg — Artifact {1} (Odyssey, #5613):
//
//	"{2}, {T}, Sacrifice this artifact: Add {B}{R}. Draw a card.
//	 (Activate only as an instant.)"
//
// One of the five Eggs, built by odyEggRow: not a mana ability, since
// it draws (CR 605.1a), so it uses the stack and an opponent may
// respond. On resolution the controller adds {B}{R}, then draws.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9b58e7fa-4259-4d6b-8b5d-33fb37fe489f",
		Name:         "Shadowblood Egg",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{odyEggRow("{B}{R}")},
	})
}
