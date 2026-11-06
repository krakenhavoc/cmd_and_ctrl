package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Syphon Fuel — Instant {4}{B} (Aetherdrift, #4464):
//
//	"Target creature gets -6/-6 until end of turn. You gain 2 life."
//
// Eject the Warp Core's kicked shrink without the kicker, then the life.
// The spell has one target, so if that creature is gone by resolution the
// spell does nothing at all, life included (CR 608.2b).
//
// No simplification.
func init() {
	purpose := ForTargets(HarmUntilEOT(0, -6, -6))
	purpose.LifeGain = 2
	Register(Spec{
		OracleID:     "8696eae4-3dc9-4572-aab9-aa4f7e8f53af",
		Name:         "Syphon Fuel",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      purpose,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if err := (BoostUntilEOT{Target: t.ID, Power: -6, Toughness: -6, Label: "Syphon Fuel"}).Apply(ctx); err != nil {
					return err
				}
			}
			return GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
		},
	})
}
