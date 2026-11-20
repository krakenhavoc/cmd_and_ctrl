package effects

// Sacred Wolf — Creature — Wolf {2}{G}, 3/1 (Magic 2011, #5579):
//
//	"Hexproof"
//
// Hexproof is the printed keyword the target check reads: its
// controller may still target it, an opponent may not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ccb64a70-cbdd-4991-a28d-b137032a4f4d",
		Name:            "Sacred Wolf",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
	})
}
