package effects

// Maned Serval — Creature — Cat {1}{W}, 1/4 (Ikoria: Lair of Behemoths, #5512):
//
//	"Vigilance"
//
// The printed keyword is the whole card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ac230f29-6d3b-4fe1-bcb0-6ffab95a1bd7",
		Name:            "Maned Serval",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
