package effects

// Twins of Maurer Estate — Creature — Vampire {4}{B}, 3/5
// (Shadows over Innistrad, #5543):
//
//	"Madness {2}{B} (If you discard this card, discard it into exile.
//	 When you do, cast it for its madness cost or put it into your
//	 graveyard.)"
//
// A body whose only text is madness. Madness is grown by the engine from
// the one string (CR 702.35a, #657).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "378110e9-62d6-4dcf-a2e3-813081c82ec4",
		Name:         "Twins of Maurer Estate",
		Completeness: CompletenessFull,
		Madness:      "{2}{B}",
	})
}
