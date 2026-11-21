package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stirge — {B} Creature — Insect Bat 1/1
//
//	"Flying
//	 This creature can't block.
//	 Blood Drain — {1}{B}, Pay 1 life, Sacrifice this creature: Draw a card."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "aed2d938-125d-49b4-93ac-202b7a37fcc2",
		Name:            "Stirge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Activated: []effects.ActivatedAbility{{
			Label:   "Blood Drain — {1}{B}, Pay 1 life, Sacrifice this creature: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.PayLife(1), effects.SacrificeThis()),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
