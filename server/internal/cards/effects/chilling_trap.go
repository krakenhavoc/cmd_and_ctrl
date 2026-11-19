package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chilling Trap — {U} Instant:
//
//	"Target creature gets -4/-0 until end of turn. If you control a
//	Wizard, draw a card."
//
// The Wizard check is made as the spell resolves (CR 608.2c).
func init() {
	Register(Spec{
		OracleID:     "7ac49129-7da2-4729-b48c-b6da279c212e",
		Name:         "Chilling Trap",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(HarmUntilEOT(0, -4, 0)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
				if err := (BoostUntilEOT{Target: t.ID, Power: -4, Label: "Chilling Trap — -4/-0"}).Apply(ctx); err != nil {
					return err
				}
			}
			if ControlsA(ctx.TextSubtype("Wizard"))(ctx.Game, ctx.Controller()) {
				return DrawCards{N: 1}.Apply(ctx)
			}
			return nil
		},
	})
}
