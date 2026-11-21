package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin War Party — Sorcery {3}{R}:
//
//	"Choose one —
//	 • Create three 1/1 red Goblin creature tokens.
//	 • Creatures you control get +1/+1 and gain haste until end of turn.
//	 Entwine {2}{R} (Choose both if you pay the entwine cost.)"
//
// Entwine (CR 702.42, #3866): {2}{R} more takes both. The bullets run in
// the order written (CR 702.42b), so entwined the three Goblins are
// already on the battlefield when the second bullet fixes its set of
// creatures (CR 611.2c) and they get +1/+1 and haste too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "fa4417e5-229a-4465-9bc5-0b3d6209b4a4",
		Name:          "Goblin War Party",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Entwine("{2}{R}")},
		Modes: effects.Entwining(effects.ChooseOne(
			effects.ModeDoing("Create three 1/1 red Goblin creature tokens.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.CreateToken{Controller: ctx.Controller(), Template: effects.RedGoblinToken(), N: 3}.Apply(ctx)
				}),
			effects.ModeDoing("Creatures you control get +1/+1 and gain haste until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					yours := effects.And(effects.Creature(), effects.YouControl())
					if err := (effects.BoostUntilEOT{Match: yours, Power: 1, Toughness: 1, Label: "Goblin War Party"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Match: yours, Keywords: []string{"haste"}, Label: "Goblin War Party — haste"}.Apply(ctx)
				}),
		)),
	})
}
