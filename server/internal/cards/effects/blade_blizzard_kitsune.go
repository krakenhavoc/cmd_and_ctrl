package effects

// Blade-Blizzard Kitsune — Creature — Fox Ninja {2}{W}, 2/2 (Neon Dynasty): "Ninjutsu {3}{W} / Double strike".
//
// Ninjutsu is the shared keyword (ninjutsu.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "84fb1c2f-cf1b-4166-834f-a887aa182d09",
		Name:            "Blade-Blizzard Kitsune",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Activated:       []ActivatedAbility{Ninjutsu("{3}{W}")},
	})
}
