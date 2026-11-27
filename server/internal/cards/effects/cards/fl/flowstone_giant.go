package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flowstone Giant — Creature — Giant {2}{R}{R}, 3/3 (Tempest, S82 OLD b756,
// #5628):
//
//	"{R}: This creature gets +2/-2 until end of turn."
//
// Breathstealer's self pump with a toughness loss: a Giant pumped past its
// toughness dies to the state-based action (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f9558fa5-2513-47e0-9af9-85db5cae4146",
		Name:         "Flowstone Giant",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gets +2/-2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: -2}},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, -2, "Flowstone Giant — +2/-2 until end of turn"),
		}},
	})
}
