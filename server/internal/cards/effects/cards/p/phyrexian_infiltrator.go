package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Infiltrator (Invasion, S82 OLD batch 647, #5616):
//
//	Creature — Phyrexian Minion {2}{B}, 2/2:
//	"{2}{U}{U}: Exchange control of this creature and target creature.
//	(This effect lasts indefinitely.)"
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f444a0f8-32d8-4748-af58-51a1a8ed72b4",
		Name:         "Phyrexian Infiltrator",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}{U}: Exchange control of this creature and target creature.",
			Cost:    effects.ManaCost("{2}{U}{U}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.S82OldB647ExchangeThisWithTarget("Phyrexian Infiltrator"),
		}},
	})
}
