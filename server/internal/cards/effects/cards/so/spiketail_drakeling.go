package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spiketail Drakeling — Creature — Drake {1}{U}{U}, 2/2 (Time Spiral,
// #5596):
//
//	"Flying
//	 Sacrifice this creature: Counter target spell unless its controller
//	 pays {2}."
//
// Spiketail Hatchling's row at {2}: the shared pay-or-be-countered body,
// asked of the spell's controller as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "989e390d-f86c-4468-a9ad-5970a571dcc7",
		Name:            "Spiketail Drakeling",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Counter target spell unless its controller pays {2}.",
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetSpell("target spell"),
			Effect:  effects.CounterTheTargetSpellUnlessPaid("Spiketail Drakeling", "{2}"),
		}},
	})
}
