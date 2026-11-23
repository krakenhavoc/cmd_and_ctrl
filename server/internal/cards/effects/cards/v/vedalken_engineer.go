package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vedalken Engineer — Creature — Vedalken Artificer {1}{U}, 1/1 (Reprint,
// #5908):
//
//	"{T}: Add two mana of any one color. Spend this mana only to cast
//	 artifact spells or activate abilities of artifacts."
//
// A mana ability (CR 605.1a) producing two of one colour (Command Chair's
// OneColorOfAmount) with Slobad's restriction: an artifact spell, or an
// ability whose source is an artifact (CR 106.6). Summoning sickness
// applies to the {T} (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b36accb4-cde6-4b1d-b549-8c975cfa999d",
		Name:         "Vedalken Engineer",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: effects.OneColorOfAmount(2),
			Label:    "{T}: Add two mana of any one color. Spend this mana only to cast artifact spells or activate abilities of artifacts.",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictType("Artifact")},
				[]string{game.ManaRestrictActivate, game.ManaRestrictType("Artifact")},
			)},
		}},
	})
}
