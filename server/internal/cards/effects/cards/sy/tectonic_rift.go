package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "b94fcdd2-7e50-439c-b5eb-aaa9ec7dee08",
		Name:         "Tectonic Rift",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DestroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			return effects.Faltering("Tectonic Rift")(item, ctx)
		},
	})
}
