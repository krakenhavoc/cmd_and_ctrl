package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reckless Reveler — Creature — Satyr {1}{R}, 2/1 (Born of the Gods,
// #5559):
//
//	"{R}, Sacrifice this creature: Destroy target artifact."
//
// Goblin Replica's ability at {R}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "17d3ceeb-c972-49bc-a530-cf67c49a7e19",
		Name:         "Reckless Reveler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, Sacrifice this creature: Destroy target artifact.",
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
