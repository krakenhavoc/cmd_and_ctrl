package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shivan Harvest — Enchantment {1}{R} (Invasion, #5616):
//
//	"{1}{R}, Sacrifice a creature: Destroy target nonbasic land."
//
// Wasteland's clause behind a creature sacrifice paid at announce (Barrage of
// Expendables' cost); "nonbasic" reads the Basic supertype off the type line.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "27c3d7cd-f92e-4203-9fb5-f6b4776f6ffd",
		Name:         "Shivan Harvest",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}, Sacrifice a creature: Destroy target nonbasic land.",
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.SacrificeACreature()),
			Targets: effects.TargetPermanent("target nonbasic land", effects.Land(), effects.B10NonbasicLand()),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
