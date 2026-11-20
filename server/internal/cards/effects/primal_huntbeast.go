package effects

// Primal Huntbeast — Creature — Beast {3}{G}, 3/3 (Magic 2013, #5568):
//
//	"Hexproof (This creature can't be the target of spells or abilities
//	 your opponents control.)"
//
// A keyword body: hexproof (CR 702.11).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "44350127-f6b6-458f-8f2f-e57a78d83303",
		Name:            "Primal Huntbeast",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
	})
}
