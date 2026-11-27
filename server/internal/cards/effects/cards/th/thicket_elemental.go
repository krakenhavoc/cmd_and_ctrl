package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thicket Elemental — Creature — Elemental {3}{G}{G}, 4/4 (Invasion, S82 OLD
// #5616):
//
//	"Kicker {1}{G} (You may pay an additional {1}{G} as you cast this spell.)
//	 When this creature enters, if it was kicked, you may reveal cards from
//	 the top of your library until you reveal a creature card. If you do, put
//	 that card onto the battlefield and shuffle all other cards revealed this
//	 way into your library."
//
// An enters trigger that exists only when kicked (CR 702.33d, 603.4). The
// "may" is asked as it resolves; yes reveals down to the first creature card
// (a token never counts, CR 108.2), puts it onto the battlefield and shuffles
// the library, which the rest never left (Shape Anew's reveal). A library with
// no creature card is revealed whole and shuffled, and nothing enters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "a85390b6-e90b-4d0c-9573-eb67ce01a0aa",
		Name:          "Thicket Elemental",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{1}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
				"Thicket Elemental — kicked, you may reveal until a creature card and put it onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Thicket Elemental — reveal until you reveal a creature card?",
						YesLabel: "Reveal",
						NoLabel:  "Don't reveal",
						Then:     thicketElementalReveal,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var thicketElementalReveal = effects.MayYes("confirm/thicket-elemental-reveal", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	return effects.OldB471RevealUntilPutOntoBattlefield(ctx, ctx.Controller(),
		func(c game.Card) bool { return c.IsCreature() }, "Thicket Elemental")
})
