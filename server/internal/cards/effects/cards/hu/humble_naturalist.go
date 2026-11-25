package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Humble Naturalist — Creature — Human Druid {1}{G}, 1/3 (Ikoria: Lair of
// Behemoths, S82 OLD batch 120, #5512):
//
//	"{T}: Add one mana of any color. Spend this mana only to cast a creature
//	 spell."
//
// Beastcaller Savant's mana ability without the haste: one mana of any color
// with the CR 106.12 spend restriction, so it pays for a creature spell and
// for nothing else. It taps, so summoning sickness gates it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff3a69a7-a244-4958-9e8b-d153942c1ba6",
		Name:         "Humble Naturalist",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			Produced:     "{W|U|B|R|G}",
			Restrictions: []string{effects.ManaRestrictCast, effects.ManaRestrictType("Creature")},
			Label:        "Add one mana of any color. Spend this mana only to cast a creature spell",
		}},
	})
}
