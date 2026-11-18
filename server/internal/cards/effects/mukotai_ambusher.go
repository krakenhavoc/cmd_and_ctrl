package effects

// Mukotai Ambusher — Artifact Creature — Rat Ninja {3}{B}, 3/2 (Neon Dynasty): "Ninjutsu {1}{B} / Lifelink".
//
// Ninjutsu is the shared keyword (ninjutsu.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3a3f0bf4-7a02-4a57-8a18-ff22f012e0ef",
		Name:            "Mukotai Ambusher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated:       []ActivatedAbility{Ninjutsu("{1}{B}")},
	})
}
