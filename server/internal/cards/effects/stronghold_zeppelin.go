package effects

// Stronghold Zeppelin — Creature — Human {2}{U}{U}, 3/3
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
		OracleID:        "dbd86a37-8ac1-4bdb-9ba1-7caee99ab5f7",
		Name:            "Stronghold Zeppelin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules:      nemBlocksOnlyFlyers(OnSelf()),
	})
}
