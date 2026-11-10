package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Faerie Fencing — Instant {X}{B} (Wilds of Eldraine, #4471):
//
//	"Target creature gets -X/-X until end of turn. That creature gets an
//	 additional -3/-3 until end of turn if you controlled a Faerie as you
//	 cast this spell."
//
// "As you cast" fixes the Faerie check when the spell is cast (CR 601.2i,
// the ruling: it no longer matters whether you still control one as the
// spell resolves), so it is a CastTimeValue. The -X/-X and the additional
// -3/-3 are applied together as one shrink.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "450a9b73-9024-4103-b9f9-216a2f395cff",
		Name:         "Faerie Fencing",
		Completeness: CompletenessFull,
		XMatters:     true,
		Purpose:      ForTargets(HarmUntilEOT(0, 0, 0)),
		Targets:      TargetCreature("target creature"),
		CastTimeValues: []game.CastTimeValue{{
			Key: "faerie",
			Value: func(g *game.Game, item *game.StackItem) int {
				return std53FaerieControlledBy(g, item.Controller)
			},
		}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			shrink := ctx.X()
			if f, _ := ctx.CastValue("faerie"); f > 0 {
				shrink += 3
			}
			if shrink <= 0 {
				return nil
			}
			return BoostUntilEOT{Target: id, Power: -shrink, Toughness: -shrink,
				Label: "Faerie Fencing — shrink until end of turn"}.Apply(ctx)
		},
	})
}
