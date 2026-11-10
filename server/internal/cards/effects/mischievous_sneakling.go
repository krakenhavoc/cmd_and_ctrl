package effects

// Mischievous Sneakling — Creature — Shapeshifter {1}{U/B} (Lorwyn Eclipsed,
// #4458):
//
//	"Changeling (This card is every creature type.)
//	 Flash"
func init() {
	Register(Spec{
		OracleID:        "fe7a4361-561a-4bab-921e-d01bab19dad3",
		Name:            "Mischievous Sneakling",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"changeling", "flash"},
	})
}
