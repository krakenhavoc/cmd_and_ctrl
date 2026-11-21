package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dross Ripper — Artifact Creature — Phyrexian Dog {4}, 3/3 (Mirrodin Besieged,
// #5576):
//
//	"{2}{B}: This creature gets +1/+1 until end of turn."
//
// Masked Blackguard's self pump (thisGetsUntilEndOfTurn).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bf3b7489-e9d6-4b3c-9fb8-3bb727584db9",
		Name:         "Dross Ripper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    effects.ManaCost("{2}{B}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Dross Ripper — +1/+1 until end of turn"),
		}},
	})
}
