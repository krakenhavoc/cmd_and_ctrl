package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clairvoyance — Instant {U} (Ice Age, S82 OLD batch 837, #5636):
//
//	"Look at target player's hand.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// Peek's look (only the caster learns the hand), then the delayed draw at the
// next upkeep at the table, whoever's turn it is (CR 603.7a,
// next_upkeep_draw.go). With its one target gone the spell does not resolve, so
// nothing is drawn (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "851333be-d7d5-472a-95cb-0a28da09ce24",
		Name:         "Clairvoyance",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			effects.LookAtTargetPlayersHand(item, ctx)
			return effects.DrawACardAtTheNextUpkeep("Clairvoyance — draw a card").Apply(ctx)
		},
	})
}
