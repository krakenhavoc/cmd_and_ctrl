package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rampant Elephant (Invasion, S82 OLD batch 647, #5616):
//
//	Creature — Elephant {3}{W}, 2/2:
//	"{G}: Target creature blocks this creature this turn if able."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aaaaf370-6486-4e17-9a93-92e05ebcf29f",
		Name:         "Rampant Elephant",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}: Target creature blocks this creature this turn if able.",
			Cost:    effects.ManaCost("{G}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.S82OldB378TargetBlocksThis,
		}},
	})
}
