package effects

// Dryad Sophisticate — Creature — Dryad {1}{G}, 2/1 (Guildpact, #5599):
//
//	"Nonbasic landwalk (This creature can't be blocked as long as
//	 defending player controls a nonbasic land.)"
//
// Trailblazer's Boots' keyword, printed on the creature: the canonical
// "nonbasic landwalk" token game/landwalk.go reads at the block check
// (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7cf59255-0ad7-4ca8-bc95-f3a8cc2643d5",
		Name:            "Dryad Sophisticate",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"nonbasic landwalk"},
	})
}
