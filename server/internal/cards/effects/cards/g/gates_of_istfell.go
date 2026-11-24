package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gates of Istfell — Land (Kaldheim, #5505):
//
//	"This land enters tapped.
//	 {T}: Add {W}.
//	 {2}{W}{U}{U}, {T}, Sacrifice this land: You gain 2 life and draw two
//	 cards."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f85c26e-3c87-4112-ad1a-8a5708555a93",
		Name:         "Gates of Istfell",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}{U}{U}, {T}, Sacrifice this land: You gain 2 life and draw two cards.",
			Purpose: game.Purpose{Draws: 2, LifeGain: 2, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{W}{U}{U}"), effects.TapCost(), effects.SacrificeThis()),
			Effect:  effects.Do(effects.GainLife{Amount: 2}, effects.DrawCards{N: 2}),
		}},
	})
}
