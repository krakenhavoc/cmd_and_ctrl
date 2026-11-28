package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Energy Bolt — Sorcery {X}{R}{W} (Mirage, #5633):
//
//	"Choose one —
//	 • Energy Bolt deals X damage to target player or planeswalker.
//	 • Target player gains X life."
//
// X is the value announced as the spell is cast (CR 107.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5c19db44-0eae-4d22-9bf4-7e6cdc6b9d8f",
		Name:         "Energy Bolt",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Energy Bolt deals X damage to target player or planeswalker.",
				effects.TargetPlayerOrPlaneswalker(),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: ctx.X()}.Apply(ctx)
				})),
			effects.ModeDoing("Target player gains X life.",
				effects.TargetPlayer("target player"),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.GainLife{Player: id, Amount: ctx.X()}.Apply(ctx)
				})),
		),
	})
}
