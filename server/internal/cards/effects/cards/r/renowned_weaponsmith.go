package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Renowned Weaponsmith — Creature — Human Artificer {1}{U}, 1/3 (Commander
// Legends: Battle for Baldur's Gate):
//
//	"{T}: Add {C}{C}. Spend this mana only to cast artifact spells or
//	 activate abilities of artifacts.
//	 {U}, {T}: Search your library for a card named Heart-Piercer Bow or
//	 Vial of Dragonfire, reveal it, put it into your hand, then shuffle."
//
// Vedalken Engineer's restricted mana ability (cast an artifact spell or
// activate an ability of an artifact) with two colourless mana, and
// Dragonstorm Forecaster's search by name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "40c1375b-1d15-4525-894a-16b7a0ac02b4",
		Name:         "Renowned Weaponsmith",
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
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}: Search your library for a card named Heart-Piercer Bow or Vial of Dragonfire, reveal it, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Mana: "{U}", Tap: true},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player: item.Controller,
					Predicate: func(c game.Card) bool {
						return game.HasName(c, "Heart-Piercer Bow") || game.HasName(c, "Vial of Dragonfire")
					},
					Dest:    game.ZoneHand,
					Limit:   1,
					Reveal:  true,
					Shuffle: true,
					Source:  item.SourceCardID,
					Reason:  "Renowned Weaponsmith — search for a card named Heart-Piercer Bow or Vial of Dragonfire",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
