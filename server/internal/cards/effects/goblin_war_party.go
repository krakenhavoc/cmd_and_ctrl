package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "fa4417e5-229a-4465-9bc5-0b3d6209b4a4",
		Name:          "Goblin War Party",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Entwine("{2}{R}")},
		Modes: Entwining(ChooseOne(
			ModeDoing("Create three 1/1 red Goblin creature tokens.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return CreateToken{Controller: ctx.Controller(), Template: RedGoblinToken(), N: 3}.Apply(ctx)
				}),
			ModeDoing("Creatures you control get +1/+1 and gain haste until end of turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					yours := And(Creature(), YouControl())
					if err := (BoostUntilEOT{Match: yours, Power: 1, Toughness: 1, Label: "Goblin War Party"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Match: yours, Keywords: []string{"haste"}, Label: "Goblin War Party — haste"}.Apply(ctx)
				}),
		)),
	})
}
