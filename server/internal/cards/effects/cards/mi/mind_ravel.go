package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mind Ravel — Sorcery {2}{B} (Ice Age, S82 OLD batch 837, #5636):
//
//	"Target player discards a card.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// The target chooses the discard (CR 701.9a). The delayed draw (CR 603.7a,
// next_upkeep_draw.go) is set up as the spell resolves, so it happens at the
// next upkeep whether or not the discard finds a card. With its one target gone
// the spell does not resolve, so nothing is drawn (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8cd0fcac-800a-431c-aca9-804998704606",
		Name:         "Mind Ravel",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.TargetPlayerDiscards(1)(ctx.Game, item); err != nil {
				return err
			}
			return effects.DrawACardAtTheNextUpkeep("Mind Ravel — draw a card").Apply(ctx)
		},
	})
}
