package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heckling Fiends — Creature — Devil {2}{R}, 2/2 (Conspiracy, S82 OLD batch
// 363, #5557):
//
//	"{2}{R}: Target creature attacks this turn if able."
//
// Imps' Taunt's attack requirement (S82OldB363TargetAttacksThisTurnIfAble),
// until end of turn: it binds only the creature's controller, and only if that
// creature can attack this turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b8fb8dc-f1c5-4994-8a58-4849e7e85539",
		Name:         "Heckling Fiends",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}: Target creature attacks this turn if able.",
			Cost:    effects.ManaCost("{2}{R}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.S82OldB363TargetAttacksThisTurnIfAble("Heckling Fiends"),
		}},
	})
}
