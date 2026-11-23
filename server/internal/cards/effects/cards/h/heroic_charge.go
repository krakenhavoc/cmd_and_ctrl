package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heroic Charge — {2}{W}{W} Instant (Dominaria United):
//
//	"Kicker {1}{R} (You may pay an additional {1}{R} as you cast this
//	 spell.)
//	 Creatures you control get +2/+1 until end of turn. If this spell was
//	 kicked, those creatures also gain trample until end of turn."
//
// The affected set is the creatures you control as it resolves (CR
// 611.2c); a creature that arrives later gets neither.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "862511a3-672f-44fe-8121-a9afe1d1404f",
		Name:          "Heroic Charge",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{1}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			yours := effects.And(effects.Creature(), effects.YouControl())
			if err := (effects.BoostUntilEOT{Match: yours, Power: 2, Toughness: 1, Label: "Heroic Charge — +2/+1"}).Apply(ctx); err != nil {
				return err
			}
			if !ctx.WasKicked() {
				return nil
			}
			return effects.GrantKeywordUntilEOT{Match: yours, Keywords: []string{"trample"}, Label: "Heroic Charge — trample"}.Apply(ctx)
		},
	})
}
