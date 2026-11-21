package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:         "2da8e9d6-e8a8-46a5-a478-3e81098d2a08",
		Name:             "Sickening Shoal",
		XMatters:         true,
		Completeness:     effects.CompletenessFull,
		Purpose:          effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
		Targets:          effects.TargetCreature("target creature"),
		AlternativeCosts: []game.AlternativeCost{effects.BokB1ShoalPitch("black", "B")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.BoostUntilEOT{Target: id, Power: -ctx.X(), Toughness: -ctx.X(),
				Label: "Sickening Shoal — -X/-X until end of turn"}.Apply(ctx)
		},
	})
}
