package effects

// Balduvian War-Makers — Creature — Human Barbarian {4}{R}, 3/3 (Alliances, #5769):
//
//	"Haste
//	 Rampage 1 (Whenever this creature becomes blocked, it gets
//	 +1/+1 until end of turn for each creature blocking it beyond the
//	 first.)"
//
// Haste is the engine's keyword, and rampage is its keyword trigger
// (CR 702.23, game/rampage.go), which counts the creatures blocking it
// as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1bda7f60-f986-455b-9d56-bbd494e27c17",
		Name:            "Balduvian War-Makers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste", "rampage 1"},
	})
}
