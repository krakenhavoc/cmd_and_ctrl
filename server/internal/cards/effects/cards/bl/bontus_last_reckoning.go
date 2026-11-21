package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bontu's Last Reckoning — Sorcery {1}{B}{B}:
//
//	"Destroy all creatures. Lands you control don't untap during your
//	 next untap step."
//
// Day of Judgment's sweep (wrathDestroyAllCreatures: simultaneous, no
// regeneration rider), then #3995's player-wide skip over the caster's
// lands, read at their next untap step: a land played before then stays
// tapped too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c010d833-1873-4132-b4ff-8d78b3d7c2d7",
		Name:         "Bontu's Last Reckoning",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.WrathDestroyAllCreatures(item, ctx); err != nil {
				return err
			}
			return effects.LandsYouControlDontUntapNextStep(ctx)
		},
	})
}
