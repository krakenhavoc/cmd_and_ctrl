package effects

import "github.com/google/uuid"

// Dazzling Lights — Instant {U} (Guilds of Ravnica, #5525):
//
//	"Target creature gets -3/-0 until end of turn.
//	 Surveil 2."
//
// Two steps: the target's -3/-0 until cleanup (CR 514.2), then surveil 2,
// which waits on its own prompt (ResolveSteps).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "394709f0-2d29-4a2d-be05-7b22fec57763",
		Name:         "Dazzling Lights",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(HarmUntilEOT(0, -3, 0)),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return BoostUntilEOT{Target: id, Power: -3, Label: "Dazzling Lights — -3/-0 until end of turn"}.Apply(ctx)
			}),
			Surveil{N: 2},
		),
	})
}
