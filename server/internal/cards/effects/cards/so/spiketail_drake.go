package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spiketail Drake — Creature — Drake {3}{U}{U}, 3/3 (Prophecy, #5617):
//
//	"Flying
//	 Sacrifice this creature: Counter target spell unless its controller
//	 pays {3}."
//
// Spiketail Hatchling's shape with a {3} tax: flying is a printed
// keyword, and the ability is counterTheTargetSpellUnlessPaid, with the
// Drake sacrificed as the cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6568859d-b96d-4015-84da-1a8960752737",
		Name:            "Spiketail Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Counter target spell unless its controller pays {3}.",
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetSpell("target spell"),
			Effect:  effects.CounterTheTargetSpellUnlessPaid("Spiketail Drake", "{3}"),
		}},
	})
}
