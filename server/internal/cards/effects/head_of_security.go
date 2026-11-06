package effects

// Head of Security — {1}{W} Creature — Andorian Soldier 2/2 (Star Trek
// Commander, #4974): "First strike." The keyword is the whole card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fb42bade-4bbc-4119-b6cc-71695fe43dd9",
		Name:            "Head of Security",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
	})
}
