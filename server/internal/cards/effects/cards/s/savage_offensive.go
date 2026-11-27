package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Savage Offensive — Sorcery {1}{R} (Invasion, S82 OLD #5616):
//
//	"Kicker {G} (You may pay an additional {G} as you cast this spell.)
//	 Creatures you control gain first strike until end of turn. If this spell
//	 was kicked, they get +1/+1 until end of turn."
//
// Bold Defense's shape: the creatures are chosen once as the spell resolves
// (CR 611.2c) and always gain first strike; kicked, those same creatures also
// get +1/+1. The +1/+1 is an addition, not an "instead".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "b5944ca8-d54e-4830-9f66-be8fe97d101d",
		Name:          "Savage Offensive",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{G}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			mine := effects.And(effects.Creature(), effects.YouControl())
			if err := (effects.GrantKeywordUntilEOT{Match: mine, Keywords: []string{"first strike"}, Label: "Savage Offensive — first strike"}).Apply(ctx); err != nil {
				return err
			}
			if !ctx.WasKicked() {
				return nil
			}
			return effects.BoostUntilEOT{Match: mine, Power: 1, Toughness: 1, Label: "Savage Offensive — +1/+1"}.Apply(ctx)
		},
	})
}
