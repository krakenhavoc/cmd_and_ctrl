package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yasharn, Implacable Earth — Legendary Creature — Elemental Boar
// {2}{G}{W}, 4/4 (#6167):
//
//	"When Yasharn enters, search your library for a basic Forest card
//	 and a basic Plains card, reveal those cards, put them into your
//	 hand, then shuffle.
//	 Players can't pay life or sacrifice nonland permanents to cast
//	 spells or activate abilities."
//
// The static is the world rule PlayersCantPayLifeOrSacrificeNonlandTo
// CastOrActivate, a second-set restriction bit Yasharn carries on itself
// (game/sacrifice_cost_bar.go): while it is on the battlefield every
// life component of a cast's or an activation's cost, mana abilities
// included, is unpayable, and no nonland permanent can pay a sacrifice
// cost (CR 101.2, CR 118.3). A land still can, so a fetch land works.
//
// The search is one prompt for up to two basic lands, held to "a basic
// Forest card and a basic Plains card" by a registered set rule; either
// may be missed (CR 701.23b), and a basic that is both is never printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6dc23410-4d8f-4c8a-92b4-a8b3fae1af4e",
		Name:         "Yasharn, Implacable Earth",
		Completeness: effects.CompletenessFull,
		// ADR 0126: the entry search finds two basic lands for the hand.
		Purpose: game.Purpose{Tutors: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Yasharn — search for a basic Forest card and a basic Plains card",
				func(g *game.Game, item *game.StackItem) error {
					forest, plains := effects.IsBasicLandWithSubtype("Forest"), effects.IsBasicLandWithSubtype("Plains")
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return forest(c) || plains(c) },
						Dest:      game.ZoneHand, Limit: 2, Reveal: true, Shuffle: true,
						Rule:   yasharnForestAndPlains,
						Source: item.SourceCardID,
						Reason: "Yasharn — a basic Forest card and a basic Plains card",
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Static: []game.StaticAbility{{
			Label:     "Players can't pay life or sacrifice nonland permanents to cast spells or activate abilities.",
			Layer:     game.Layer6Ability,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.ExtraRestrictions |= game.PlayersCantPayLifeOrSacrificeNonlandToCastOrActivate
			},
		}},
	})
}

// yasharnForestAndPlains holds the search to at most one basic Forest
// and one basic Plains card.
var yasharnForestAndPlains = game.RegisterSearchRule("search/yasharn-forest-and-plains", func(picked []game.Card, _ int) bool {
	forests, plains := 0, 0
	for _, c := range picked {
		switch {
		case c.HasSubtypeWord("Forest"):
			forests++
		case c.HasSubtypeWord("Plains"):
			plains++
		}
	}
	return len(picked) <= 2 && forests <= 1 && plains <= 1 && forests+plains == len(picked)
})
