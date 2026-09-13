package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Razaketh, the Foulblooded — Legendary Creature — Demon {5}{B}{B}{B},
// 8/8 (EDHREC rank 1861):
//
//	"Flying, trample
//	 Pay 2 life, Sacrifice another creature: Search your library for
//	 a card, put that card into your hand, then shuffle."
//
// The Demon that turns every creature into a Demonic Tutor. The
// cost is two components merged by Plus — life and a sacrifice of
// another creature — and the tutor is unrestricted: the S22 searcher
// chooses from the whole library.
//
// "Another creature" is the Warren Soultrader shape: the sacrifice
// clause excludes the source by NAME, since a cost's spec is built at
// init and has no instance to compare against. Razaketh is legendary,
// so a second one can never be on the battlefield to be wrongly
// refused.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "136c9ecb-59b4-4ef6-bcb2-7b8d3df3ee75",
		Name:            "Razaketh, the Foulblooded",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Activated: []ActivatedAbility{{
			Label: "Pay 2 life, Sacrifice another creature: Search your library for a card and put it into your hand",
			Cost: Plus(PayLife(2), game.AbilityCost{
				SacrificeOther: sacrificeSpec("another creature", Creature(), b03NotNamed("Razaketh, the Foulblooded")),
			}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(game.Card) bool { return true },
					Dest:      game.ZoneHand,
					Limit:     1,
					Shuffle:   true,
					Reason:    "Razaketh, the Foulblooded — a card, to hand",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
