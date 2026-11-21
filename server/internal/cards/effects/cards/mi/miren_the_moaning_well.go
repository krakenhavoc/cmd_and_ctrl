package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Miren, the Moaning Well — Legendary Land:
//
//	"{T}: Add {C}.
//	 {3}, {T}, Sacrifice a creature: You gain life equal to the
//	 sacrificed creature's toughness."
//
// Greater Good's sacrifice cost and last-known-information read, with
// toughness (CR 608.2h) and a life gain. Counters and an anthem's bonus
// count, because the toughness is the one the creature had on the
// battlefield. A creature whose toughness was zero or less gains
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "03fe19bb-8e22-4030-8299-2ddd2d5a7eb2",
		Name:         "Miren, the Moaning Well",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice a creature: You gain life equal to the sacrificed creature's toughness.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.SacrificeACreature()),
			Effect:  effects.GainLifeEqualToSacrificedToughness,
		}},
	})
}
