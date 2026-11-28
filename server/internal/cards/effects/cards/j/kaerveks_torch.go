package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kaervek's Torch — Sorcery {X}{R} (Mirage, #5633):
//
//	"As long as Kaervek's Torch is on the stack, spells that target it cost {2}
//	 more to cast.
//	 Kaervek's Torch deals X damage to any target."
//
// The first line is an ability of the spell that functions on the stack
// (CR 113.6b, #6222): every spell that targets the Torch costs {2} more, whoever
// casts it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "c629feae-084c-452f-9c09-4b9716331498",
		Name:          "Kaervek's Torch",
		Completeness:  effects.CompletenessFull,
		XMatters:      true,
		CostModifiers: []game.CostModifier{effects.SpellsThatTargetThisCostMore(2, "Kaervek's Torch — spells that target it cost {2} more to cast")},
		Targets:       effects.TargetAny(),
		Purpose:       effects.ForTargets(effects.DamageXToTarget(0)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CmdrB19DealXToTheTarget(ctx)
		},
	})
}
