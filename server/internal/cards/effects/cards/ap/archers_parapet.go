package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Archers' Parapet — Khans of Tarkir (S82 OLD b348, #5554):
//
//	Creature — Wall {1}{G}, 0/5: Defender. {1}{B}, {T}: Each opponent loses 1 life.

func init() {
	effects.Register(effects.Spec{
		OracleID:        "9df1c48b-52a8-47d7-8315-272fecf1a8a0",
		Name:            "Archers' Parapet",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, {T}: Each opponent loses 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost()),
			Effect:  effects.Do(effects.EachOpponentLosesLife{N: 1}),
		}},
	})
}
