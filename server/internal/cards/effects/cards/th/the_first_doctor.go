package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The First Doctor — Legendary Creature — Time Lord Doctor {1}{W}{U}, 2/2
// (Doctor Who, #5948):
//
//	"When The First Doctor enters, search your library and/or graveyard for a
//	 card named TARDIS, reveal it, and put it into your hand. If you search
//	 your library this way, shuffle.
//	 Whenever you cast a spell with cascade, put a +1/+1 counter on target
//	 artifact or creature."
//
// The search is Delivery Moogle's (SearchLibrary over the library and the
// graveyard, revealing the card and shuffling when the library was
// searched). The second ability is a cast trigger over a spell that has
// cascade, read off the spell's own cascade trigger row; its target is an
// artifact or creature of any controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5611b70d-f8d1-4a8f-b9a2-a356264a7c46",
		Name:         "The First Doctor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("The First Doctor — search your library and/or graveyard for a card named TARDIS, reveal it, and put it into your hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:        item.Controller,
						Predicate:     func(c game.Card) bool { return c.Name == "TARDIS" },
						Dest:          game.ZoneHand,
						Limit:         1,
						Reveal:        true,
						Shuffle:       true,
						AlsoGraveyard: true,
						Reason:        "The First Doctor — a card named TARDIS",
					}.Apply(effects.NewContext(g, item))
				}),
			effects.Targeting(
				effects.WheneverYouCast(effects.GapB218SpellWithCascade, "The First Doctor — put a +1/+1 counter on target artifact or creature",
					effects.B36CounterOnChosenAnimal),
				effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature()))),
		},
	})
}
