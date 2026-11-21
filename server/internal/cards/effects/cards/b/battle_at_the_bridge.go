package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battle at the Bridge — Sorcery {X}{B}: "Improvise ... Target creature
// gets -X/-X until end of turn. You gain X life."
//
// Improvise is Spec.TapCost; artifacts pay the generic {X}, not the {B}.
// If the target is gone the whole spell fizzles, so no life is gained.
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d63ad3a-4e9f-4968-a576-4be063f4ecbf",
		Name:         "Battle at the Bridge",
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		TapCost:      effects.Improvise(),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			if err := effects.VsPumpTheTarget(ctx, -x, -x, "Battle at the Bridge — -X/-X"); err != nil {
				return err
			}
			return effects.GainLife{Player: ctx.Controller(), Amount: x}.Apply(ctx)
		},
	})
}
