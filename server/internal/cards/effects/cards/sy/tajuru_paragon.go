package cardssy

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tajuru Paragon — Creature — Elf {1}{G}, 3/2 (Zendikar Rising, S82 OLD
// #5508):
//
//	"Tajuru Paragon is also a Cleric, Rogue, Warrior, and Wizard.
//	 Kicker {3}
//	 When this creature enters, if it was kicked, reveal the top six cards
//	 of your library. You may put a card that shares a creature type with it
//	 from among them into your hand. Put the rest on the bottom of your
//	 library in a random order."
//
// The extra types are a layer 4 effect on the Paragon itself, so it counts
// toward the party as any of them (CR 700.8). "Shares a creature type with
// it" is read against the Paragon's creature types as the trigger
// resolves (CR 205.3d); a changeling shares every type.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "b03a9d03-4853-4856-8cfa-ad4085e0e3e3",
		Name:          "Tajuru Paragon",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{3}")},
		Static:        []game.StaticAbility{effects.S82OldB090AlsoAPartyCreature("Tajuru Paragon is also a Cleric, Rogue, Warrior, and Wizard.")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
				"Tajuru Paragon — kicked: reveal the top six cards; you may put a card that shares a creature type with it into your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Controller()
					paragon, _ := g.LookupCardForEffect(item.SourceCardID)
					types := paragon.Effective().Subtypes
					return effects.TakeFromLibraryToHand{
						Player: player,
						Cards:  g.RevealTopOfLibraryForEffect(player, ctx.Source(), 6, "Tajuru Paragon — reveal the top six cards"),
						Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
							return slices.ContainsFunc(types, c.HasSubtype)
						},
						Max:      1,
						Optional: true,
						Label:    "Tajuru Paragon — you may put a card that shares a creature type with it into your hand",
						Key:      effects.TakeRevealRestBottomRandom,
					}.Apply(ctx)
				}),
		},
	})
}
