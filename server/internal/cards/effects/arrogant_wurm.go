package effects

// Arrogant Wurm — Creature — Wurm {3}{G}{G}, 4/4 (Torment, #5612):
//
//	"Trample
//	 Madness {2}{G} (If you discard this card, discard it into exile.
//	 When you do, cast it for its madness cost or put it into your
//	 graveyard.)"
//
// Trample is a printed keyword (CR 702.19b), and madness is the engine's
// (CR 702.35a, #657), grown from the one string: a discarded Wurm is
// cast from exile for {2}{G}, at instant speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dc7f0c33-80ea-453e-a237-473f077f038e",
		Name:            "Arrogant Wurm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Madness:         "{2}{G}",
	})
}
