package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stonybrook Angler — Creature — Merfolk Wizard {1}{U}, 1/2 (Jumpstart 2022, #5948):
//
//	"{1}{U}, {T}: You may tap or untap target creature."
//
// Bounding Krasis's shared "you may tap or untap the target" body behind an
// activated ability: the choice is made as the ability resolves, and a
// target that left is illegal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d29d28c8-4ba9-4fcc-b0d7-5d7b3de89b82",
		Name:         "Stonybrook Angler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}, {T}: You may tap or untap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.NemMayTapOrUntapTheTarget("Stonybrook Angler"),
		}},
	})
}
