package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sinew Dancer — Creature — Phyrexian Soldier {W}, 1/1:
//
//	"{3}{W}, {T}: Tap target creature.
//	 Corrupted — {W}, {T}: Tap target creature. Activate only if an opponent
//	 has three or more poison counters."
//
// Two separate abilities. The cheaper one carries the corrupted activation
// condition (CR 602.1b), read as it is announced.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a605864a-f46f-4ebc-8e56-ce7322c918b0",
		Name:         "Sinew Dancer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}{W}, {T}: Tap target creature.",
				Cost:    effects.Plus(effects.ManaCost("{3}{W}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.ThousandMoonsCrackshotTap,
			},
			{
				Label:     "Corrupted — {W}, {T}: Tap target creature. Activate only if an opponent has three or more poison counters.",
				Cost:      effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
				Targets:   effects.TargetCreature("target creature"),
				Condition: effects.CorruptedAnd(nil),
				Effect:    effects.ThousandMoonsCrackshotTap,
			},
		},
	})
}
