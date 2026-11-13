package effects

// Rootwater Commando — Creature — Merfolk {2}{U}, 2/2 (Nemesis, #5618):
//
//	"Islandwalk (This creature can't be blocked as long as defending
//	 player controls an Island.)"
//
// Cold-Eyed Selkie's islandwalk without the draw: a printed keyword,
// enforced by the block check (game/landwalk.go, CR 702.14c). An Island
// of any player but the defending one doesn't stop a block.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d62e9510-dc41-493f-8289-6e11303884fc",
		Name:            "Rootwater Commando",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
	})
}
