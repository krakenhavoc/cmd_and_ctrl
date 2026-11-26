package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fulminator Mage — Creature — Elemental Shaman {1}{B/R}{B/R}, 2/2 (Shadowmoor, #5591):
//
//	"Sacrifice this creature: Destroy target nonbasic land."
//
// Wasteland's nonbasic-land target, paid for by sacrificing the creature
// at announce, so it is gone before the ability resolves (CR 602.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd4c46a3-b723-4b35-9061-9a1dee7cc9d8",
		Name:         "Fulminator Mage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Destroy target nonbasic land.",
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetPermanent("target nonbasic land", effects.Land(), effects.B10NonbasicLand()),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
