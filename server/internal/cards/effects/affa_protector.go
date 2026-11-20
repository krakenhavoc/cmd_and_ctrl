package effects

// Affa Protector — Creature — Human Soldier Ally {2}{W}, 1/4 (Oath of the Gatewatch, #5545):
//
//	"Vigilance"
//
// Its whole text is a printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5b38a1e4-a6af-4094-9392-70a87759e20d",
		Name:            "Affa Protector",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
