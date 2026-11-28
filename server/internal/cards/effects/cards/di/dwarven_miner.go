package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dwarven Miner — Creature — Dwarf {1}{R}, 1/2 (Mirage):
//
//	"{2}{R}, {T}: Destroy target nonbasic land."
//
// Wasteland's nonbasic-land target with a mana-and-tap cost on a creature, so
// summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0935377b-384f-4e4c-9fbb-8d2a7d5dc280",
		Name:         "Dwarven Miner",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, {T}: Destroy target nonbasic land.",
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target nonbasic land", effects.Land(), effects.B10NonbasicLand()),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
