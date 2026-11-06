package effects

// Deer-Dog — Creature — Elk Dog {1}{R}, 1/3 (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"First strike (This creature deals combat damage before creatures
//	 without first strike.)"
//
// First strike is a printed keyword (CR 702.7) the combat damage step
// already reads.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "76c7a339-015b-4074-8e10-14d52b784f9b",
		Name:            "Deer-Dog",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
	})
}
