package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "c4f7cdb7-aef9-4ab8-995b-4e5180d47214",
		Name:         "Sarkhan's Triumph",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.B06TutorToHand("Sarkhan's Triumph — a Dragon creature card", effects.B36IsDragonCreatureCard),
	})
}
