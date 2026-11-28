package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Folk of An-Havva — Creature — Human {G}, 1/1 (Homelands, S82 OLD b823, #5635):
//
//	"Whenever this creature blocks, it gets +2/+0 until end of turn."
//
// Giant Badger's block trigger (CR 509.3a, once however many attackers it
// blocks) with a +2/+0 pump that ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "86002526-ed3d-47e5-9f2a-55c4fa220ce8",
		Name:         "Folk of An-Havva",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.SelfBlocksOnce(ev, source)
			}, "Folk of An-Havva — gets +2/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(2, 0, "Folk of An-Havva — +2/+0")),
		},
	})
}
