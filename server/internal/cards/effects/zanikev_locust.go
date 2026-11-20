package effects

// Zanikev Locust — Creature — Insect {5}{B}, 3/3 (Return to Ravnica, #5566):
//
//	"Flying / Scavenge {2}{B}{B} ({2}{B}{B}, Exile this card from your
//	 graveyard: Put a number of +1/+1 counters equal to this card's
//	 power on target creature. Scavenge only as a sorcery.)"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
// Scavenge is the shared graveyard ability (Deadbridge Goliath,
// scavenge.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e29012bb-4a1d-409e-8db9-50920fdf68a2",
		Name:            "Zanikev Locust",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []ActivatedAbility{Scavenge("{2}{B}{B}")},
	})
}
