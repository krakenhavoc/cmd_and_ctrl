package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Errant Doomsayers — Creature — Human Rebel {1}{W}, 1/1 (Time Spiral,
// #5596):
//
//	"{T}: Tap target creature with toughness 2 or less."
//
// Cephalid Retainer's tapper with a toughness bound, judged at announce
// and again as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "82d4d70b-afb9-4ae7-8f30-16bc64dccf3e",
		Name:         "Errant Doomsayers",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Tap target creature with toughness 2 or less.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature with toughness 2 or less", effects.ToughnessLE(2)),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
