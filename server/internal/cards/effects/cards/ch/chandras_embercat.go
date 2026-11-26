package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chandra's Embercat — Creature — Elemental Cat {1}{R}, 2/2 (Core Set 2020):
//
//	"{T}: Add {R}. Spend this mana only to cast an Elemental spell or a
//	 Chandra planeswalker spell."
//
// A mana ability (CR 605.1a) with Flamebraider's restriction narrowed to casts
// (CR 106.6): one tag admits a cast of an Elemental spell, or a cast of a
// planeswalker spell that is a Chandra. It pays for an activated ability of
// nothing, and for no other spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e75bec5c-538d-44ce-a6c7-85580d04fac8",
		Name:         "Chandra's Embercat",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "{T}: Add {R}. Spend this mana only to cast an Elemental spell or a Chandra planeswalker spell.",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictSubtype("Elemental")},
				[]string{game.ManaRestrictCast, game.ManaRestrictType("Planeswalker"), game.ManaRestrictSubtype("Chandra")},
			)},
		}},
	})
}
