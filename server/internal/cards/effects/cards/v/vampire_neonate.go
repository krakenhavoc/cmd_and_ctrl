package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vampire Neonate — Creature — Vampire {B}, 0/3 (Foundations, #4465):
//
//	"{2}, {T}: Each opponent loses 1 life and you gain 1 life."
//
// Acolyte of Aclazotz's drain with a mana-and-tap cost: every opponent
// loses 1 life, then you gain 1, however many opponents there are. The
// row has no target and can be activated at instant speed, so it
// declares that it answers nothing (ADR 0142).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c0045925-25bf-403c-b721-5f05ba30985b",
		Name:         "Vampire Neonate",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Each opponent loses 1 life and you gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue, LifeGain: 1},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Effect:  effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1}),
		}},
	})
}
