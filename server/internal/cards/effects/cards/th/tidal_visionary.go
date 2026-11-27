package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tidal Visionary (Invasion, S82 OLD batch 646, #5616):
//
//	Creature — Merfolk Wizard {U}, 1/1:
//	"{T}: Target creature becomes the color of your choice until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8fc72080-8d4d-427d-ae30-065bc8ee9a28",
		Name:         "Tidal Visionary",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature becomes the color of your choice until end of turn.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.S82OldB646TargetBecomesChosenColor("Tidal Visionary"),
		}},
	})
}
