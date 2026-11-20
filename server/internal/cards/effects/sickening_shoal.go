package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sickening Shoal — Instant — Arcane {X}{B}{B} (Betrayers of Kamigawa,
// #5602):
//
//	"You may exile a black card with mana value X from your hand rather
//	 than pay this spell's mana cost.
//	 Target creature gets -X/-X until end of turn."
//
// The pitch is Shining Shoal's (bokB1ShoalPitch): X is announced and the
// exiled black card's mana value must equal it. Hard-cast, X is paid as
// usual.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "2da8e9d6-e8a8-46a5-a478-3e81098d2a08",
		Name:             "Sickening Shoal",
		XMatters:         true,
		Completeness:     CompletenessFull,
		Purpose:          ForTargets(HarmUntilEOT(0, 0, 0)),
		Targets:          TargetCreature("target creature"),
		AlternativeCosts: []game.AlternativeCost{bokB1ShoalPitch("black", "B")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return BoostUntilEOT{Target: id, Power: -ctx.X(), Toughness: -ctx.X(),
				Label: "Sickening Shoal — -X/-X until end of turn"}.Apply(ctx)
		},
	})
}
