package effects

// Belbe's Percher — Creature — Bird {2}{B}, 2/2
// (Nemesis, #5618):
//
//	"Flying
//	 This creature can block only creatures with flying."
//
// Flying is printed data (CR 702.9b). "Can block only creatures with
// flying" is Cloud Spirit's block rule (nemBlocksOnlyFlyers), judged as
// blockers are declared (CR 509.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "81b51bb1-89b9-46a0-ad7a-a43c0d214f11",
		Name:            "Belbe's Percher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules:      nemBlocksOnlyFlyers(OnSelf()),
	})
}
