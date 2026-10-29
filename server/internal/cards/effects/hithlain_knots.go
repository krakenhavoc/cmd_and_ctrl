package effects

import "github.com/google/uuid"

// Hithlain Knots — Instant {1}{U} (The Hobbit Eternal, HOC):
//
//	"Tap target creature. Scry 1.
//	 Draw a card."
//
// The two sentences are one spell resolving in order (CR 608.2c): the tap,
// the scry, then the draw, which waits for the scry's answer (#4454).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2c80fc55-86a0-403e-82ce-09ee03095d8a",
		Name:         "Hithlain Knots",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error { return TapTarget{Target: id}.Apply(ctx) }),
			Scry{N: 1},
			DrawCards{N: 1},
		),
	})
}
