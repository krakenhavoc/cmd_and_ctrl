package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Havoc Festival — Enchantment {4}{B}{R}:
//
//	"Players can't gain life.
//	 At the beginning of each player's upkeep, that player loses half
//	 their life, rounded up."
//
// "Players can't gain life" is ADR 0107 §5's battlefield static
// (CR 119.7, #1880), its controller included. The upkeep trigger fires
// on every seat's upkeep; "that player" is the active player the event
// names, and the halving reads their life as the trigger resolves
// (playerLosesHalfTheirLife: 21 loses 11, 1 loses 1).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "af0dd100-67f0-4d7c-a65b-b2c7149288f7",
		Name:         "Havoc Festival",
		Completeness: CompletenessFull,
		CantGainLife: PlayersCantGainLife(),
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil
			}, "Havoc Festival — that player loses half their life", func(g *game.Game, item *game.StackItem) error {
				return playerLosesHalfTheirLife(g, item, triggeringActor(item))
			}),
		},
	})
}
