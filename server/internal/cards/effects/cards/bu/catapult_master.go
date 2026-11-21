package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Catapult Master — Creature — Human Soldier {3}{W}{W}, 3/3 (Onslaught):
//
//	"Tap five untapped Soldiers you control: Exile target creature."
//
// Ancestor's Prophet's tap-five cost over Soldiers (tapFiveUntapped):
// the Master may be one of the five, and with no {T} symbol in the cost
// a Soldier that arrived this turn may pay it (CR 302.6). The exile is
// Hanged Executioner's body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "02da2145-0edb-44be-8622-360182a8c25a",
		Name:         "Catapult Master",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap five untapped Soldiers you control: Exile target creature.",
			Cost:    effects.TapFiveUntapped("five untapped Soldiers you control", "Soldier"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.ExileFirstLegalCardTarget,
		}},
	})
}
