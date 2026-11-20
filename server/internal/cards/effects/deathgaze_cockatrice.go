package effects

// Deathgaze Cockatrice — Creature — Cockatrice {2}{B}{B}, 2/2 (Magic 2014, #5563):
//
//	"Flying
//	 Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)"
//
// Midnight Assassin's two keywords on a 2/2 flier.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0fabb6e4-35cf-4e56-a27c-ff627d8e9b7a",
		Name:            "Deathgaze Cockatrice",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
	})
}
