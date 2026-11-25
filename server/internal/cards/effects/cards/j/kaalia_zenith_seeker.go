package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kaalia, Zenith Seeker — Legendary Creature — Human Cleric {R}{W}{B}, 3/3
// (Core Set 2020):
//
//	"Flying, vigilance
//	 When Kaalia enters, look at the top six cards of your library. You may
//	 reveal an Angel card, a Demon card, and/or a Dragon card from among them
//	 and put them into your hand. Put the rest on the bottom of your library in
//	 a random order."
//
// Flying and vigilance are printed keywords. The enters trigger is In the
// Presence of Ages' two-slot take with three slots, one card each: the three
// cards are one prompt whose set rule lets each card fill one slot it fits and
// no slot twice (an Angel Dragon fills only one), so an Angel Demon Dragon
// can't give three cards. The cards taken are revealed (CR 701.20a), and the
// rest go to the bottom in a random order. Nothing need be taken.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6307ec34-bef6-4ac6-a363-f9bef0d03b3b",
		Name:            "Kaalia, Zenith Seeker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Kaalia — look at the top six cards; you may reveal an Angel card, a Demon card, and/or a Dragon card and put them into your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Controller()
					return effects.TakeFromLibraryToHand{
						Player:   player,
						Cards:    g.LookAtTopOfLibraryForEffect(player, 6),
						Optional: true,
						Label:    "Kaalia, Zenith Seeker — you may reveal an Angel card, a Demon card, and/or a Dragon card and put them into your hand",
						Key:      kaaliaZenithSeekerTake,
					}.Apply(ctx)
				}),
		},
	})
}

// kaaliaZenithSeekerTake is the look's registered rule and rest: three slots of
// one card each, revealed, the rest on the bottom in a random order. An
// on-disk key: never renamed, never reused.
var kaaliaZenithSeekerTake = effects.RegisterTake("choose-cards/kaalia-zenith-seeker-take", effects.TakeDef{
	Reveal: true,
	Slots: []effects.TakeSlot{
		{Label: "an Angel card", Match: effects.OfSubtype("Angel"), Max: 1},
		{Label: "a Demon card", Match: effects.OfSubtype("Demon"), Max: 1},
		{Label: "a Dragon card", Match: effects.OfSubtype("Dragon"), Max: 1},
	},
	Then: effects.TakeRestOnBottomInRandomOrder,
})
