package effects

// Moldering Karok — Creature — Zombie Crocodile {2}{B}{G}, 3/3
// (Strixhaven: School of Mages, #5502):
//
//	"Trample, lifelink"
//
// Two printed keywords the combat engine reads (Startled Relic Sloth's
// pair).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5d95db27-53bb-4a20-9fb8-24998b03e767",
		Name:            "Moldering Karok",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "lifelink"},
	})
}
