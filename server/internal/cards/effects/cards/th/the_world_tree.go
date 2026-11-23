package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The World Tree — Land:
//
//	"This land enters tapped.
//	 {T}: Add {G}.
//	 As long as you control six or more lands, lands you control have
//	 "{T}: Add one mana of any color."
//	 {W}{W}{U}{U}{B}{B}{R}{R}{G}{G}, {T}, Sacrifice this land: Search
//	 your library for any number of God cards, put them onto the
//	 battlefield, then shuffle."
//
// The grant is ADR 0093's layer-6 grant with its condition in the
// "applies to": with five lands it applies to nothing, with six it
// applies to every land you control, the Tree included. The land count
// is re-read on every layer pass, and every land entering or leaving
// bumps the layer version.
//
// The God search is an unbounded search ("any number of", the Ugin
// shape) straight onto the battlefield.
//
// No simplification.
const theWorldTreeGrant = "the-world-tree/any-color"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "3437d504-bf62-4c27-b15f-f6330182ff7e",
		Name:         "The World Tree",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Grants: []effects.AbilityGrant{effects.AnyColorManaGrant(theWorldTreeGrant)},
		Static: []game.StaticAbility{effects.GrantAbilities(func(target *game.Card, g *game.Game, source *game.Card) bool {
			return effects.LandsYouControl(target, g, source) && theWorldTreeLandCount(g, source) >= 6
		}, theWorldTreeGrant)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}{W}{U}{U}{B}{B}{R}{R}{G}{G}, {T}, Sacrifice this land: Search for any number of God cards",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{W}{W}{U}{U}{B}{B}{R}{R}{G}{G}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player: item.Controller,
					// SearchLibrary lends the source's text (CR 612.2, #3984).
					Predicate: func(c game.Card) bool { return c.HasSubtypeWord("God") },
					Dest:      game.ZoneBattlefield,
					Unbounded: true,
					Shuffle:   true,
					Reason:    "The World Tree — search for any number of God cards",
					Source:    item.SourceCardID,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// theWorldTreeLandCount is how many lands the Tree's controller
// controls.
func theWorldTreeLandCount(g *game.Game, source *game.Card) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsLand() && c.Controller == source.Controller {
			n++
		}
	}
	return n
}
