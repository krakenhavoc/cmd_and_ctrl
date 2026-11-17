package effects

// Copper Host Crusher — Creature — Phyrexian Bear Rhino 8/8 {6}{G}{G} (March of the Machine, S82):
//
//	"Trample
//	 Hexproof"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6971eac1-afd0-46c3-a37f-4a26d50ef656",
		Name:            "Copper Host Crusher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "hexproof"},
	})
}
