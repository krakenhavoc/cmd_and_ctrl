package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kaya's Guile — Instant {1}{W}{B}:
//
//	"Choose two —
//	 • Each opponent sacrifices a creature of their choice.
//	 • Exile all opponents' graveyards.
//	 • Create a 1/1 white and black Spirit creature token with flying.
//	 • You gain 4 life.
//	 Entwine {3} (Choose all if you pay the entwine cost.)"
//
// The one entwine card that prints "choose two" (CR 702.42, #3866):
// unentwined it takes exactly two bullets, entwined all four — never
// three. Entwining raises the count to every bullet, minimum and
// maximum together.
//
// The edict is Pick Your Poison's (EachPlayerSacrifices, each opponent
// choosing their own); the graveyard exile walks each opponent's pile as
// Soul-Guide Lantern does. Each bullet is its own body, run in printed
// order (CR 608.2c, CR 702.42b), and the later bullets wait for the
// edict's prompts (#2789).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "a6becd5f-d279-4df8-81d9-deeb1b513c95",
		Name:          "Kaya's Guile",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Entwine("{3}")},
		Modes: Entwining(ChooseN("Choose two", 2, 2,
			ModeDoing("Each opponent sacrifices a creature of their choice.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return EachPlayerSacrifices{ExceptController: true, Match: Creature(), Label: "a creature"}.Apply(ctx)
				}),
			ModeDoing("Exile all opponents' graveyards.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					for _, opp := range ctx.Opponents() {
						if err := exileGraveyardForEffect(ctx.Game, item, opp); err != nil {
							return err
						}
					}
					return nil
				}),
			ModeDoing("Create a 1/1 white and black Spirit creature token with flying.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return CreateToken{Controller: item.Controller, Template: kayasGuileSpirit(), N: 1}.Apply(ctx)
				}),
			ModeDoing("You gain 4 life.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return GainLife{Player: item.Controller, Amount: 4}.Apply(ctx)
				}),
		)),
	})
}

// kayasGuileSpirit is the printed token: a 1/1 white and black Spirit
// with flying.
func kayasGuileSpirit() game.Card {
	return game.Card{
		Name:      "Spirit",
		TypeLine:  "Token Creature — Spirit",
		Power:     1,
		Toughness: 1,
		Colors:    []string{"W", "B"},
		Keywords:  []string{"flying"},
	}
}
