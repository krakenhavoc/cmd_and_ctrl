package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trench Wurm — Creature — Wurm {3}{B}, 3/3 (Invasion, #5616):
//
//	"{2}{R}, {T}: Destroy target nonbasic land."
//
// Wasteland's clause with a mana-and-tap cost; the {T} means the Wurm can't
// use it the turn it arrives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9cf65178-6408-46da-a940-f5cb0960b61f",
		Name:         "Trench Wurm",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, {T}: Destroy target nonbasic land.",
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target nonbasic land", effects.Land(), effects.B10NonbasicLand()),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
