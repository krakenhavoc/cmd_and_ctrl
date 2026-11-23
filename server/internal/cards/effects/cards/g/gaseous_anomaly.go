package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gaseous Anomaly — Land (Star Trek, #4973):
//
//	"{T}: Add {C}.
//	 {5}{U}, {T}, Sacrifice this land: Draw two cards. Create a tapped
//	 Planet land token named New Planet with "{T}: Add one mana of any
//	 color.""
//
// Artificial Anomaly's frame with a draw two: the tap, the mana and the
// sacrifice are costs (CR 602.2, 118.3), and the New Planet is the
// printed token (ttrk 27).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1fb736bf-c9bd-45ab-86ce-7f138f1724b2",
		Name:         "Gaseous Anomaly",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{U}, {T}, Sacrifice this land: Draw two cards. Create a tapped New Planet.",
			Purpose: game.Purpose{Draws: 2, Tokens: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{5}{U}"), effects.TapCost(), effects.SacrificeThis()),
			Effect:  effects.Do(effects.DrawCards{N: 2}, effects.TrekB28TappedNewPlanet()),
		}},
	})
}
