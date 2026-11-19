package effects

// Blor the Impervious — {3}{G}{G} Legendary Creature — Frog Warrior 3/3
// (Mystery Booster Commander, S82 GAP batch 3, #5948):
//
//	"Trample, hexproof, indestructible"
//
// Three printed keywords, read through the layered view (CR 702.19,
// 702.11, 702.12).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f5f2fa96-8e87-44bc-8c82-e217706d3cab",
		Name:            "Blor the Impervious",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "hexproof", "indestructible"},
	})
}
