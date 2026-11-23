package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fortress Cyclops — Creature — Cyclops Soldier {3}{R}{W}, 3/3 (Gatecrash, #5565):
//
//	"Whenever this creature attacks, it gets +3/+0 until end of turn.
//	 Whenever this creature blocks, it gets +0/+3 until end of turn."
//
// Flowstone Charger's attack trigger and Royal Trooper's block trigger,
// each pinned to this creature until cleanup.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7790bd6c-75e2-4abd-87ab-e1c15f1d53ab",
		Name:         "Fortress Cyclops",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Fortress Cyclops — it gets +3/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(3, 0, "Fortress Cyclops — +3/+0")),
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.SelfBlocksOnce(ev, source)
			}, "Fortress Cyclops — it gets +0/+3 until end of turn",
				effects.ThisGetsUntilEndOfTurn(0, 3, "Fortress Cyclops — +0/+3")),
		},
	})
}
