package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fire of Orthanc — Sorcery {3}{R} (The Hobbit Eternal, HOC):
//
//	"Destroy target artifact or land. Creatures without flying can't
//	 block this turn."
//
// Two sentences the catalog already has (#4454): the destroy is the shared
// destroyTheTargetPermanent over an artifact-or-land target clause, and the
// second sentence is Falter's rule, faltering(), a restriction on a SET read
// live until cleanup rather than a list fixed as the spell resolves
// (CR 611.2c). If the target is gone as the spell resolves, the whole spell
// does nothing (CR 608.2b), so no creature is stopped from blocking.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "407021e2-4072-4961-95be-f9309c341677",
		Name:         "Fire of Orthanc",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact or land", effects.Or(effects.Artifact(), effects.Land())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DestroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			return effects.Faltering("Fire of Orthanc")(item, ctx)
		},
	})
}
