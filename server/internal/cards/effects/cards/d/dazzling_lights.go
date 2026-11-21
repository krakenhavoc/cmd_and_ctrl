package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "394709f0-2d29-4a2d-be05-7b22fec57763",
		Name:         "Dazzling Lights",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -3, 0)),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.BoostUntilEOT{Target: id, Power: -3, Label: "Dazzling Lights — -3/-0 until end of turn"}.Apply(ctx)
			}),
			effects.Surveil{N: 2},
		),
	})
}
