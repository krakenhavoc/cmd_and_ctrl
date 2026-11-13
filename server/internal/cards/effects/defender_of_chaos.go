package effects

// Defender of Chaos — Creature — Human Knight {2}{R}, 2/1 (Urza's
// Legacy, #5623):
//
//	"Flash
//	 Protection from white"
//
// Both are printed keywords: flash lets it be cast any time an instant could
// be (CR 702.8a), and protection from white is read by targeting, blocking
// and damage (CR 702.16a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e5f04e8a-e9ae-4f00-8ada-f63b3454ed30",
		Name:            "Defender of Chaos",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "protection from white"},
	})
}
