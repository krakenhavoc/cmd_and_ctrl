package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grasp of Darkness — Instant {B}{B} (Scars of Mirrodin, #5577):
//
//	"Target creature gets -4/-4 until end of turn."
//
// Disfigure with a bigger number.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "494c819d-7d73-432b-8c96-4cb9dcd31094",
		Name:         "Grasp of Darkness",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(HarmUntilEOT(0, -4, -4)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return BoostUntilEOT{Target: id, Power: -4, Toughness: -4, Label: "Grasp of Darkness — -4/-4"}.Apply(ctx)
		},
	})
}
