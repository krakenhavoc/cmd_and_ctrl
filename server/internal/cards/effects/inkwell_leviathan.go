package effects

// Inkwell Leviathan — Artifact Creature — Leviathan {7}{U}{U}, 7/11:
//
//	"Trample
//	 Islandwalk (This creature can't be blocked as long as defending
//	 player controls an Island.)
//	 Shroud (This creature can't be the target of spells or
//	 abilities.)"
//
// All three keywords are enforced by the engine (trample overflow,
// landwalk in block legality, shroud at the targeting choke point), so
// the card is three printed keywords and nothing else. Shroud stops the
// controller's own spells and abilities too, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dbdd0963-b395-4cb8-ad9a-ed85a3e9f2e5",
		Name:            "Inkwell Leviathan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "islandwalk", "shroud"},
	})
}
