package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hargilde, Kindly Runechanter — Legendary Creature — Human {2}{W}{U}, 2/3
// (Secret Lair Drop):
//
//	"{T}: Add {C}{C}. Spend this mana only to cast artifact spells or activate
//	 abilities of artifacts.
//	 Partner—Friends forever (You can have two commanders if both have this
//	 ability.)"
//
// A mana ability (CR 605.1a) adding {C}{C} with Vedalken Engineer's restriction:
// an artifact spell, or an ability whose source is an artifact (CR 106.6).
// Summoning sickness applies to the {T} (CR 302.6). Partner—Friends forever is
// a deck-construction rule read off the oracle text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "747009e7-da07-48c7-b838-687f19b884ea",
		Name:         "Hargilde, Kindly Runechanter",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "{T}: Add {C}{C}. Spend this mana only to cast artifact spells or activate abilities of artifacts.",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictType("Artifact")},
				[]string{game.ManaRestrictActivate, game.ManaRestrictType("Artifact")},
			)},
		}},
	})
}
