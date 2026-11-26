package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Manaforge Cinder — Creature — Elemental Shaman {B/R}, 1/1 (Shadowmoor,
// #5591):
//
//	"{1}: Add {B} or {R}. Activate no more than three times each turn."
//
// A mana ability filtering {1} into one black or red mana. The limit is
// AtMostNActivationsPerTurn: the per-object count of activations this turn
// (CR 602.1b), which every mana activation writes before it is paid.
//
// No simplification.
func init() {
	const label = "{1}: Add {B} or {R}. Activate no more than three times each turn."
	effects.Register(effects.Spec{
		OracleID:     "3c9535a2-7995-42de-a56f-0a445f2dd697",
		Name:         "Manaforge Cinder",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:      effects.ManaAbilityCost{Mana: "{1}"},
			Produced:  "{B|R}",
			Label:     label,
			Condition: effects.AtMostNActivationsPerTurn(label, 3),
		}},
	})
}
