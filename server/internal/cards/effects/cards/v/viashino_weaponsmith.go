package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viashino Weaponsmith — Creature — Lizard {3}{R} 2/2 (Urza's Saga, #5624):
//
//	"Whenever this creature becomes blocked by a creature, this creature gets
//	 +2/+2 until end of turn."
//
// One trigger per blocker (EventBlock names each blocker), as the printed
// "by a creature" reads.
//
// No simplification.
func init() {
	const label = "Viashino Weaponsmith — this creature gets +2/+2 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "f4d17f41-efd6-4173-ab6e-28b6fbcee10d",
		Name:         "Viashino Weaponsmith",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, effects.ViciousBattleragerBlocked, label, effects.ThisGetsUntilEndOfTurn(2, 2, label)),
		},
	})
}
