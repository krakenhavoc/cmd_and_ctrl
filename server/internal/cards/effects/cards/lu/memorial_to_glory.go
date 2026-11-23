package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Memorial to Glory — Land:
//
//	"This land enters tapped.
//	 {T}: Add {W}.
//	 {3}{W}, {T}, Sacrifice this land: Create two 1/1 white Soldier
//	 creature tokens."
//
// The white Memorial: a tapped Plains that makes two Soldiers late.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eb8ec34c-ae07-4a09-940f-ee965146a787",
		Name:         "Memorial to Glory",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}, {T}, Sacrifice this land: Create two 1/1 white Soldier creature tokens.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}{W}"), effects.TapCost(), effects.SacrificeThis()),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 2}),
		}},
	})
}
