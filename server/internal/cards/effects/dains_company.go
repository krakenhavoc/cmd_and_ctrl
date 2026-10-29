package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dáin's Company — Creature — Dwarf Warrior {R}{W}, 2/2 (The Hobbit, HOB):
//
//	"This creature has lifelink as long as you control another Dwarf.
//	 When this creature enters, look at the top four cards of your library.
//	 You may reveal a Dwarf or Equipment card from among them and put it
//	 into your hand. Put the rest on the bottom of your library in a random
//	 order."
//
// The lifelink is Bolg's Company's haste: a self-only layer 6 keyword grant
// gated on some OTHER Dwarf its controller controls, read live. The look is
// Monumental Henge's: a choose_cards prompt over the four cards with a
// "may" floor of zero, the taken card revealed, and the rest to the bottom
// in a random order (takeRevealRestBottomRandom).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ae03782f-8f05-4ce2-a049-005b1b688faf",
		Name:         "Dáin's Company",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(hobA9WhileYouControlAnother("Dwarf"), "lifelink"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dáin's Company — look at the top four cards; you may reveal a Dwarf or Equipment card and put it into your hand, the rest on the bottom in a random order",
				func(g *game.Game, item *game.StackItem) error {
					player := item.Controller
					return TakeFromLibraryToHand{
						Player:   player,
						Cards:    g.LookAtTopOfLibraryForEffect(player, 4),
						Match:    Or(Subtype("Dwarf"), Subtype("Equipment")),
						Max:      1,
						Optional: true,
						Label:    "Dáin's Company — you may reveal a Dwarf or Equipment card and put it into your hand",
						Key:      takeRevealRestBottomRandom,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
