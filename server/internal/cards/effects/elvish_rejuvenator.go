package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elvish Rejuvenator — Creature — Elf Druid {2}{G}, 1/1 (Outlaws of Thunder
// Junction Commander, #5368):
//
//	"When this creature enters, look at the top five cards of your library. You
//	 may put a land card from among them onto the battlefield tapped. Put the
//	 rest on the bottom of your library in a random order."
//
// Aang, at the Crossroads' look-and-put with a land filter and the printed
// "tapped": a private look (CR 701.20), one optional land card from among
// them, the rest on the bottom in a random order whether or not one was put.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "71d14755-b81d-4aec-b94d-1637884a5844",
		Name:         "Elvish Rejuvenator",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Elvish Rejuvenator — look at the top five cards, you may put a land card from among them onto the battlefield tapped, then the rest on the bottom in a random order",
				func(g *game.Game, item *game.StackItem) error {
					return PutFromLibraryOntoBattlefield{
						Player:   item.Controller,
						Cards:    g.LookAtTopOfLibraryForEffect(item.Controller, 5),
						Match:    Land(),
						Max:      1,
						Optional: true,
						Tapped:   true,
						Label:    "Elvish Rejuvenator — you may put a land card from among them onto the battlefield tapped",
						ThenTail: PutRestOnBottomInRandomOrderTail,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
