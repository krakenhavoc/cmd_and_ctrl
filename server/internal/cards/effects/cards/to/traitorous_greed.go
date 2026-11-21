package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Traitorous Greed — #4380, S76. Uses the existing catalog seams.
func init() {
	effects.Register(effects.Spec{
		OracleID: "344cdec9-60b3-42ec-a3d5-41ee1d4f7c6c", Name: "Traitorous Greed", Completeness: effects.CompletenessFull,
		Targets: effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.GainControl{Target: id, Duration: effects.DurationUntilEndOfTurn(ctx), Label: "Traitorous Greed — gain control"}.Apply(ctx)
			}),
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error { return effects.UntapTarget{Target: id}.Apply(ctx) }),
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Traitorous Greed — haste"}.Apply(ctx)
			}),
			effects.AddMana{Produced: effects.OneColorOfAmount(2)},
		),
	})
}
