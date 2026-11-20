package effects

// Kraul Raider — Creature — Insect Warrior {2}{B}, 2/3 (Guilds of Ravnica, #5525):
//
//	"Menace"
//
// Menace is an engine keyword carried as printed data (CR 702.111b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3c741801-bfa5-4ff6-bf06-8d5edc06388d",
		Name:            "Kraul Raider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
	})
}
