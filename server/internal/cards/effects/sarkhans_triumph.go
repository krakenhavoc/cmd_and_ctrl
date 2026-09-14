package effects

// Sarkhan's Triumph — Instant {2}{R} (EDHREC rank 3851):
//
//	"Search your library for a Dragon creature card, reveal it, put it
//	 into your hand, then shuffle."
//
// The Dragon tutor. Instant-speed, type-filtered search to hand,
// revealed, through the S22 chooser — Eladamri's Call's body with a
// Dragon creature filter on the printed type line (effective
// subtypes are not readable in a library).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c4f7cdb7-aef9-4ab8-995b-4e5180d47214",
		Name:         "Sarkhan's Triumph",
		Completeness: CompletenessFull,
		OnResolve:    b06TutorToHand("Sarkhan's Triumph — a Dragon creature card", b36IsDragonCreatureCard),
	})
}
