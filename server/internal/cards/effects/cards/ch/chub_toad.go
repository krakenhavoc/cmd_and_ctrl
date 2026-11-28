package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chub Toad — Creature — Frog {2}{G}, 1/1 (Ice Age, S82 OLD, #5636):
//
//	"Whenever this creature blocks or becomes blocked, it gets +2/+2 until end of
//	 turn."
//
// Brushwagg's trigger (one for either event) with +2/+2.
//
// No simplification.
func init() {
	const label = "Chub Toad — gets +2/+2 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "a0daace2-92fc-4c7d-b497-94b4ba775559",
		Name:         "Chub Toad",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventBlock, game.EventBecomesBlocked}, effects.ThisBlocksOrBecomesBlocked,
				label, effects.ThisGetsUntilEndOfTurn(2, 2, label)),
		},
	})
}
