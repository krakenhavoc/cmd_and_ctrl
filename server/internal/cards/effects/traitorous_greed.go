package effects

import "github.com/google/uuid"

// Traitorous Greed — #4380, S76. Uses the existing catalog seams.
func init() {
	Register(Spec{
		OracleID: "344cdec9-60b3-42ec-a3d5-41ee1d4f7c6c", Name: "Traitorous Greed", Completeness: CompletenessFull,
		Targets: TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return GainControl{Target: id, Duration: DurationUntilEndOfTurn(ctx), Label: "Traitorous Greed — gain control"}.Apply(ctx)
			}),
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error { return UntapTarget{Target: id}.Apply(ctx) }),
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Traitorous Greed — haste"}.Apply(ctx)
			}),
			AddMana{Produced: OneColorOfAmount(2)},
		),
	})
}
