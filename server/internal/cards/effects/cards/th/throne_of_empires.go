package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Throne of Empires — Artifact {4} (Magic 2012, S82 older sets batch 451,
// #5573):
//
//	"{1}, {T}: Create a 1/1 white Soldier creature token. Create five of
//	 those tokens instead if you control artifacts named Crown of Empires
//	 and Scepter of Empires."
//
// The count is decided as the ability resolves (CR 608.2c): an artifact of
// each name, read by name through the layered view.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "12afe8e6-eaf0-45c9-8086-4c658db7cb7e",
		Name:         "Throne of Empires",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Create a 1/1 white Soldier creature token. Create five of those tokens instead if you control artifacts named Crown of Empires and Scepter of Empires.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Effect:  effects.S82OldB451ThroneOfEmpires,
		}},
	})
}
