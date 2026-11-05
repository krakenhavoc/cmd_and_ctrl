package effects

// Sickbay Orderly — {2}{W} Creature — Caitian Doctor 3/2 (Star Trek
// Commander, #4974): "Lifelink." The keyword is the whole card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "94688e62-f39f-41fd-9309-f16b715ef724",
		Name:            "Sickbay Orderly",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
