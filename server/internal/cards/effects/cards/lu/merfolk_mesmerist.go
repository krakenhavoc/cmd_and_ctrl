package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Merfolk Mesmerist — Creature — Merfolk Wizard {1}{U}, 1/2 (Magic 2012,
// #5573):
//
//	"{U}, {T}: Target player mills two cards."
//
// Millstone's row at {U}, {T}: the target player mills two cards, or what
// is left of their library (CR 701.17b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "590d524f-278d-4cf4-ae3f-efcced4517b1",
		Name:         "Merfolk Mesmerist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}: Target player mills two cards.",
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.B31MillChosenPlayer(2),
		}},
	})
}
