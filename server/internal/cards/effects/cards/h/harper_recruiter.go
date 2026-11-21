package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harper Recruiter — Creature — Human Warrior {2}{W}, 3/1 (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Flying
//	 Whenever this creature attacks, look at the top four cards of your
//	 library. You may reveal a Cleric card, a Rogue card, a Warrior card,
//	 and/or a Wizard card from among them and put those cards into your hand.
//	 Put the rest on the bottom of your library in a random order."
//
// Gift of the Gargantuan's look with one slot per class: a card that is two
// of the classes (a Cleric Warrior) fills only one slot, and each slot
// holds one card. The cards taken are revealed; the rest go to the bottom
// in a random order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9ecbc7de-9399-42f4-9c39-b9255080f48c",
		Name:            "Harper Recruiter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Harper Recruiter — look at the top four cards; you may reveal a Cleric, a Rogue, a Warrior and/or a Wizard card and put them into your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Controller()
					return effects.TakeFromLibraryToHand{
						Player:   player,
						Cards:    g.LookAtTopOfLibraryForEffect(player, 4),
						Optional: true,
						Label:    "Harper Recruiter — reveal a Cleric card, a Rogue card, a Warrior card and/or a Wizard card and put them into your hand",
						Key:      harperRecruiterTake,
					}.Apply(ctx)
				}),
		},
	})
}

// harperRecruiterTake is the look's registered rule and rest: four slots of
// one card each, revealed, then the rest on the bottom in a random order.
// An on-disk key: never renamed, never reused.
var harperRecruiterTake = effects.RegisterTake("choose-cards/harper-recruiter-take",
	effects.TakeDef{
		Reveal: true,
		Slots: []effects.TakeSlot{
			{Label: "a Cleric card", Match: effects.Subtype("Cleric"), Max: 1},
			{Label: "a Rogue card", Match: effects.Subtype("Rogue"), Max: 1},
			{Label: "a Warrior card", Match: effects.Subtype("Warrior"), Max: 1},
			{Label: "a Wizard card", Match: effects.Subtype("Wizard"), Max: 1},
		},
		Then: effects.TakeRestOnBottomInRandomOrder,
	})
