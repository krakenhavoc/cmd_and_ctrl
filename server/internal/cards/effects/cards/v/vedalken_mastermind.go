package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vedalken Mastermind — Creature — Vedalken Wizard {U}{U}, 1/2 (Fifth Dawn, #5604):
//
//	"{U}, {T}: Return target permanent you control to its owner's hand."
//
// Rescue's bounce of a permanent you control, as a {U}, {T} ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ba85537-6dc2-459e-8850-619b41c0a881",
		Name:         "Vedalken Mastermind",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}: Return target permanent you control to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target permanent you control", effects.YouControl()),
			Effect:  effects.BounceTheTargetToHand,
		}},
	})
}
