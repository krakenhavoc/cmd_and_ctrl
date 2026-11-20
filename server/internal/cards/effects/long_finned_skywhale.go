package effects

// Long-Finned Skywhale — Creature — Whale {2}{U}{U}, 4/3 (Kaladesh, #5539):
//
//	"Flying
//	 This creature can block only creatures with flying."
//
// Flying is printed; the block restriction is Battlefield Percher's
// (nemBlocksOnlyFlyers), judged as blockers are declared (CR 509.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a8287efd-23ca-498b-90f1-c76099919fa4",
		Name:            "Long-Finned Skywhale",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules:      nemBlocksOnlyFlyers(OnSelf()),
	})
}
