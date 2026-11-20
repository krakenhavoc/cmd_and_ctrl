package effects

// Sluiceway Scorpion — Creature — Scorpion {2}{B}{G}, 2/2 (Return to Ravnica, #5566):
//
//	"Deathtouch (Any amount of damage this deals to a creature is
//	 enough to destroy it.) / Scavenge {1}{B}{G} ({1}{B}{G}, Exile this
//	 card from your graveyard: Put a number of +1/+1 counters equal to
//	 this card's power on target creature. Scavenge only as a
//	 sorcery.)"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
// Scavenge is the shared graveyard ability (Deadbridge Goliath,
// scavenge.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bb5ea7c7-741a-4edd-9696-b9881b70afe0",
		Name:            "Sluiceway Scorpion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated:       []ActivatedAbility{Scavenge("{1}{B}{G}")},
	})
}
