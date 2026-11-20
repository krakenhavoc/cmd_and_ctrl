package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tectonic Rift — Sorcery {3}{R} (Magic 2012, #5573):
//
//	"Destroy target land. Creatures without flying can't block this
//	 turn."
//
// Stone Rain's destroy, then Falter's restriction (faltering): the
// creatures without flying as it resolves can't block this turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b94fcdd2-7e50-439c-b5eb-aaa9ec7dee08",
		Name:         "Tectonic Rift",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target land", Land()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := destroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			return faltering("Tectonic Rift")(item, ctx)
		},
	})
}
