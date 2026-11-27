package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cursecatcher — Creature — Merfolk Wizard {U}, 1/1 (Shadowmoor, #5591):
//
//	"Sacrifice this creature: Counter target instant or sorcery spell unless
//	 its controller pays {1}."
//
// Spiketail Drakeling's pay-or-be-countered body at {1}, aimed at an
// instant or sorcery spell. The sacrifice is paid at announce, so the
// Cursecatcher is gone before the ability resolves; the pay-unless is legal
// against a spell that can't be countered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a59a23f2-158f-4dea-8593-75fe7b31b726",
		Name:         "Cursecatcher",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Counter target instant or sorcery spell unless its controller pays {1}.",
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetInstantOrSorcerySpell("target instant or sorcery spell"),
			Effect:  effects.CounterTheTargetSpellUnlessPaid("Cursecatcher", "{1}"),
		}},
	})
}
