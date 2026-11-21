package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voyaging Satyr — Creature — Satyr Druid {1}{G}, 1/2 (Theros, #5561):
//
//	"{T}: Untap target land."
//
// Portent Tracker's untap row: any land, whoever controls it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bce18f34-848e-4dd2-9a38-6084c0695f78",
		Name:         "Voyaging Satyr",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Untap target land.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target land", effects.Land()),
			Effect:  effects.B22UntapEachLegalTarget,
		}},
	})
}
