package effects

import "github.com/google/uuid"

// Throttle — Instant {4}{B} (Khans of Tarkir, #5554):
//
//	"Target creature gets -4/-4 until end of turn."
//
// Grasp of Darkness's -4/-4 at a different cost: the shrink ends in the
// cleanup step (CR 514.2), and a target that left in response gets
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "01823ab7-b9f6-48bc-b45c-9b0fabc6f657",
		Name:         "Throttle",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(HarmUntilEOT(0, -4, -4)),
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
			return BoostUntilEOT{Target: id, Power: -4, Toughness: -4, Label: "Throttle — -4/-4"}.Apply(ctx)
		})),
	})
}
