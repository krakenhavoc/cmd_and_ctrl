package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vedalken Entrancer — Creature — Vedalken Wizard {3}{U}, 1/4 (Jumpstart,
// #5510):
//
//	"{U}, {T}: Target player mills two cards."
//
// Millstone's mill body at {U}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "04cc5292-a485-4bb8-a143-9364229d480f",
		Name:         "Vedalken Entrancer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}: Target player mills two cards.",
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.B31MillChosenPlayer(2),
		}},
	})
}
