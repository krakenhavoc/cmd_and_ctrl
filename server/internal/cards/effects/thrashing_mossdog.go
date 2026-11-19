package effects

// Thrashing Mossdog — Creature — Plant Dog {3}{G}, 3/3 (Ravnica: Clue Edition, #5929): "Reach. Scavenge {4}{G}{G}."
//
// Scavenge is the shared graveyard ability (Deadbridge Goliath).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8778ddbe-9aba-4a45-a22a-5453fefef952",
		Name:            "Thrashing Mossdog",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated:       []ActivatedAbility{Scavenge("{4}{G}{G}")},
	})
}
