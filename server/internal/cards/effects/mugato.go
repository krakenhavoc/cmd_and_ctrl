package effects

// Mugato — {3}{G}{G} Creature — Ape Beast, 6/6 (Star Trek, #4973):
// "Reach, vigilance, deathtouch".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "eb9a3b52-62ad-467f-b62a-0273ceee48e5",
		Name:            "Mugato",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance", "deathtouch"},
	})
}
