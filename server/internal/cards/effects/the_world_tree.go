package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "3437d504-bf62-4c27-b15f-f6330182ff7e",
		Name:         "The World Tree",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Grants: []AbilityGrant{AnyColorManaGrant(theWorldTreeGrant)},
		Static: []game.StaticAbility{GrantAbilities(func(target *game.Card, g *game.Game, source *game.Card) bool {
			return landsYouControl(target, g, source) && theWorldTreeLandCount(g, source) >= 6
		}, theWorldTreeGrant)},
		Activated: []ActivatedAbility{{
			Label: "{W}{W}{U}{U}{B}{B}{R}{R}{G}{G}, {T}, Sacrifice this land: Search for any number of God cards",
			Cost:  Plus(ManaCost("{W}{W}{U}{U}{B}{B}{R}{R}{G}{G}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.HasSubtype("God") },
					Dest:      game.ZoneBattlefield,
					Unbounded: true,
					Shuffle:   true,
					Reason:    "The World Tree — search for any number of God cards",
					Source:    item.SourceCardID,
				}.Apply(NewContext(g, item))
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
