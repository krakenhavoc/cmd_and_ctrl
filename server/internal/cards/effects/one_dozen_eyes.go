package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// One Dozen Eyes — Sorcery {5}{G}:
//
//	"Choose one —
//	 • Create a 5/5 green Beast creature token.
//	 • Create five 1/1 green Insect creature tokens.
//	 Entwine {G}{G}{G} (Choose both if you pay the entwine cost.)"
//
// Two token bullets; entwined (CR 702.42, #3866) the Beast is made
// first, then the five Insects (CR 702.42b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "b1fbf818-6699-4f05-9a91-19aa296526bf",
		Name:          "One Dozen Eyes",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Entwine("{G}{G}{G}")},
		Modes: Entwining(ChooseOne(
			ModeDoing("Create a 5/5 green Beast creature token.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return CreateToken{Controller: item.Controller, Template: oneDozenEyesBeast(), N: 1}.Apply(ctx)
				}),
			ModeDoing("Create five 1/1 green Insect creature tokens.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 green Insect"), N: 5}.Apply(ctx)
				}),
		)),
	})
}

// oneDozenEyesBeast is the printed 5/5 green Beast token.
func oneDozenEyesBeast() game.Card {
	return game.Card{
		Name:      "Beast",
		TypeLine:  "Token Creature — Beast",
		Power:     5,
		Toughness: 5,
		Colors:    []string{"G"},
	}
}
