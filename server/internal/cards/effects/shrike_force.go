package effects

// Shrike Force — Creature — Bird Knight {2}{W}, 1/3 (Bloomburrow, #4467):
//
//	"Flying, double strike, vigilance"
//
// Serra Angel's shape: three printed keywords and nothing else, fed to
// the layer 6 keyword static through PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d014b02f-a8b8-4b97-96e0-f588792069f0",
		Name:            "Shrike Force",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "double strike", "vigilance"},
	})
}
