package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "7790bd6c-75e2-4abd-87ab-e1c15f1d53ab",
		Name:         "Fortress Cyclops",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Fortress Cyclops — it gets +3/+0 until end of turn",
				thisGetsUntilEndOfTurn(3, 0, "Fortress Cyclops — +3/+0")),
			On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return selfBlocksOnce(ev, source)
			}, "Fortress Cyclops — it gets +0/+3 until end of turn",
				thisGetsUntilEndOfTurn(0, 3, "Fortress Cyclops — +0/+3")),
		},
	})
}
