package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Life's Finale — Sorcery {4}{B}{B}:
//
//	"Destroy all creatures, then search target opponent's library for up to
//	 three creature cards and put them into their graveyard. Then that player
//	 shuffles."
//
// The destruction is the ordinary wrath; the search is the spell's
// controller's, over the chosen opponent's library, and may find fewer.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c1d5e6b-f6b6-4dc6-9bc1-3cc00b78da36",
		Name:         "Life's Finale",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			victim, ok := effects.FirstLegalPlayerTarget(ctx)
			if err := effects.WrathDestroyAllCreatures(item, ctx); err != nil {
				return err
			}
			if !ok {
				return nil
			}
			return effects.LifesFinaleSearch(ctx, victim)
		},
	})
}
