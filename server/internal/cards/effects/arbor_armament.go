package effects

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arbor Armament — #4380, S76. Uses the existing catalog seams.
func init() {
	Register(Spec{
		OracleID: "504f21e8-82ac-4433-b5e9-640989480f58", Name: "Arbor Armament", Completeness: CompletenessFull,
		Targets: TargetCreature("target creature"), Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "reach")),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"reach"}, Label: "Arbor Armament — reach"}.Apply(ctx)
			}),
		),
	})
}
