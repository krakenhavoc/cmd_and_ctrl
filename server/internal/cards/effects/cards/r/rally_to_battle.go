package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rally to Battle — Instant {3}{W} (Ravnica: Clue Edition, #5929):
//
//	"Creatures you control get +1/+3 until end of turn. Untap them."
//
// The boost is fixed on the creatures you control as the spell resolves
// (CR 611.2c) and "them" is that same set, untapped right after.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d8d88dd-c266-4b7d-a756-ac2d8976ca89",
		Name:         "Rally to Battle",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 3, Label: "Rally to Battle — +1/+3"}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error { return effects.UntapAllCreaturesYouControl{}.Apply(ctx) }),
		),
	})
}
