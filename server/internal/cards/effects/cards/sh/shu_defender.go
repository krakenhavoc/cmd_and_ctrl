package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shu Defender — Creature — Human Soldier {2}{W}, 2/2
// (Portal Three Kingdoms, #5622):
//
//	"Whenever this creature blocks, it gets +0/+2 until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b6625cb-4177-45bd-ba99-2af43dfd6f46",
		Name:         "Shu Defender",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.SelfBlocksOnce(ev, source)
			}, "Shu Defender — gets +0/+2 until end of turn",
				effects.ThisGetsUntilEndOfTurn(0, 2, "Shu Defender — +0/+2")),
		},
	})
}
