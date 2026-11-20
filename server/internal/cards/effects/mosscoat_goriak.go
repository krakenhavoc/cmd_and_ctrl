package effects

// Mosscoat Goriak — Creature — Beast {2}{G}, 2/4 (Ikoria: Lair of Behemoths, #5512):
//
//	"Vigilance"
//
// The printed keyword is the whole card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "28cc015f-1f3e-417a-9762-de25109fee5c",
		Name:            "Mosscoat Goriak",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
