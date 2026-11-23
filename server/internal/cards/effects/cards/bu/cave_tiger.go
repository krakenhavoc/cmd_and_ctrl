package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cave Tiger — Creature — Cat {2}{G}, 2/2 (Urza's Saga):
//
//	"Whenever this creature becomes blocked by a creature, this creature
//	 gets +1/+1 until end of turn."
//
// The trigger is per blocker (CR 509.3d): EventBlock is one event per
// blocker–attacker pair, Vicious Battlerager's shape, so a Tiger blocked
// by two creatures gets +2/+2. The Tiger blocking is not it becoming
// blocked, and does not trigger it.
//
// No simplification.
func init() {
	const label = "Cave Tiger — +1/+1 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "d3bac8b8-0b2a-4c67-b087-7a0bf1359468",
		Name:         "Cave Tiger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, effects.ViciousBattleragerBlocked, label, effects.ThisCreatureUntilEOT(label, 1, 1)),
		},
	})
}
