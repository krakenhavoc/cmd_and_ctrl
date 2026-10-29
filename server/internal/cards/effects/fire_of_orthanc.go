package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "407021e2-4072-4961-95be-f9309c341677",
		Name:         "Fire of Orthanc",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact or land", Or(Artifact(), Land())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := destroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			return faltering("Fire of Orthanc")(item, ctx)
		},
	})
}
