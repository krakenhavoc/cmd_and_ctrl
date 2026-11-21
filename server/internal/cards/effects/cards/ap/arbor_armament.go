package cardsap

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arbor Armament — #4380, S76. Uses the existing catalog seams.
func init() {
	effects.Register(effects.Spec{
		OracleID: "504f21e8-82ac-4433-b5e9-640989480f58", Name: "Arbor Armament", Completeness: effects.CompletenessFull,
		Targets: effects.TargetCreature("target creature"), Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "reach")),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"reach"}, Label: "Arbor Armament — reach"}.Apply(ctx)
			}),
		),
	})
}
