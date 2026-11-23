package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barbarian Riftcutter — Creature — Human Barbarian {4}{R}, 3/3
// (Ravnica: City of Guilds, #5600):
//
//	"{R}, Sacrifice this creature: Destroy target land."
//
// Strip Mine's destroy on a creature's sacrifice: the creature is paid as
// the cost, so the ability resolves with its source already gone, and a
// land that left in response is not destroyed (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9612e4c-1527-4255-9f2f-b8d8cee65cc6",
		Name:         "Barbarian Riftcutter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, Sacrifice this creature: Destroy target land.",
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target land", effects.Land()),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
