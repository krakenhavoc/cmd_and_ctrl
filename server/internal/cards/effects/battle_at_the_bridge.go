package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Battle at the Bridge — Sorcery {X}{B}: "Improvise ... Target creature
// gets -X/-X until end of turn. You gain X life."
//
// Improvise is Spec.TapCost; artifacts pay the generic {X}, not the {B}.
// If the target is gone the whole spell fizzles, so no life is gained.
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3d63ad3a-4e9f-4968-a576-4be063f4ecbf",
		Name:         "Battle at the Bridge",
		Completeness: CompletenessFull,
		XMatters:     true,
		TapCost:      Improvise(),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := ctx.X()
			if err := vsPumpTheTarget(ctx, -x, -x, "Battle at the Bridge — -X/-X"); err != nil {
				return err
			}
			return GainLife{Player: ctx.Controller(), Amount: x}.Apply(ctx)
		},
	})
}
