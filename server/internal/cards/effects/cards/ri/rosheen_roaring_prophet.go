package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rosheen, Roaring Prophet — Legendary Creature — Giant Shaman
// {2}{R}{G}, 4/4 (#6085, S82):
//
//	"When Rosheen enters, mill six cards. You may put a card with {X} in
//	 its mana cost from among them into your hand.
//	 {T}: Reveal any number of cards with {X} in their mana cost in your
//	 hand. Add {C}{C} for each card revealed this way. Spend this mana
//	 only on costs that contain {X}."
//
// The enters trigger is Scout the City's mill-then-take over cards with
// {X} in their mana cost (game.CardHasXInManaCost). The mana ability is
// #6085's reveal (RevealAnyNumber): the activator names any number of
// matching cards in hand, none included; they are revealed as the ability
// resolves (CR 605.3b, CR 701.20a) and stay in the hand (CR 701.20b), and
// each adds {C}{C} under Elementalist's Palette's restriction
// (game.ManaRestrictXCost, CR 106.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3108a4fe-b18a-4d02-9545-b19812454362",
		Name:         "Rosheen, Roaring Prophet",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rosheen, Roaring Prophet — mill six cards; you may put a card with {X} in its mana cost from among them into your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillToZone{N: 6, ThenTail: ctx.ItemTail(rosheenMilled)}.Apply(ctx)
				}),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{
				Tap:    true,
				Reveal: effects.RevealAnyNumber("cards with {X} in their mana cost", game.RevealCost{XInManaCost: true}),
			},
			ProducedForPaid: rosheenProduced,
			Restrictions:    []string{game.ManaRestrictXCost},
			Label:           "{T}: Reveal any number of cards with {X} in their mana cost in your hand. Add {C}{C} for each card revealed this way. Spend this mana only on costs that contain {X}.",
		}},
	})
}

var rosheenMilled = effects.CardTail("rosheen-roaring-prophet/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.MayTakeOneFromAmongThem(ctx, o.IDs, func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return game.CardHasXInManaCost(c)
	}, "Rosheen, Roaring Prophet — you may put a card with {X} in its mana cost from among them into your hand")
})

// rosheenProduced is "Add {C}{C} for each card revealed this way".
func rosheenProduced(_ *game.Game, _, _ uuid.UUID, paid game.PaidCost) game.ManaOutput {
	return game.ManaOutput{Amount: game.ManaAmount{Options: []string{"C"}, N: 2 * len(paid.Revealed)}}
}
