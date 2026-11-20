package effects

// Skysnare Spider — Creature — Spider {4}{G}{G}, 6/6 (Magic Origins, #5549):
//
//	"Vigilance (Attacking doesn't cause this creature to tap.)
//	 Reach (This creature can block creatures with flying.)"
//
// Keywords only: vigilance and reach.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b820c29e-2428-4074-a024-9502848484d6",
		Name:            "Skysnare Spider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "reach"},
	})
}
