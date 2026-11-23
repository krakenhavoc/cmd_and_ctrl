package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Artificial Anomaly — Land (Star Trek, #4973):
//
//	"{T}: Add {C}.
//	 {3}{B}, {T}, Sacrifice this land: Each opponent loses 2 life and you
//	 gain 2 life. Create a tapped Planet land token named New Planet with
//	 "{T}: Add one mana of any color.""
//
// The tap, the mana and the sacrifice are all costs (CR 602.2, 118.3).
// The drain is a fixed 2 either way: the gain does not depend on what the
// opponents lost. The New Planet is the printed token (ttrk 27).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e550295-eaf4-4962-bc52-261d7c2d71ae",
		Name:         "Artificial Anomaly",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{B}, {T}, Sacrifice this land: Each opponent loses 2 life and you gain 2 life. Create a tapped New Planet.",
			Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}{B}"), effects.TapCost(), effects.SacrificeThis()),
			Effect:  effects.Do(effects.EachOpponentLosesLife{N: 2}, effects.GainLife{Amount: 2}, effects.TrekB28TappedNewPlanet()),
		}},
	})
}
