package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rebuild — Instant {2}{U} (Urza's Legacy, #5623):
//
//	"Return all artifacts to their owners' hands.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Evacuation for artifacts: every artifact returns to its OWNER's hand,
// whoever controls it, and a token artifact ceases to exist there. Cycling is
// the shared row (CR 702.29a). Declared as a bounce sweep of artifacts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "496999a7-a82a-45ea-9c4e-1b6aabc781cd",
		Name:         "Rebuild",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifacts, How: game.SweepBounce}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BounceAllMatching{Match: Artifact()}.Apply(ctx)
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
