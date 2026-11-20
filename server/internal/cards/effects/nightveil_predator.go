package effects

// Nightveil Predator — Creature — Vampire {U}{U}{B}{B}, 3/3 (Guilds of Ravnica, #5525):
//
//	"Flying, deathtouch
//	 Hexproof"
//
// Three engine keywords carried as printed data: flying, deathtouch and
// hexproof (CR 702.11b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b94591b8-2e99-4983-8f36-9b52207bfd2e",
		Name:            "Nightveil Predator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch", "hexproof"},
	})
}
