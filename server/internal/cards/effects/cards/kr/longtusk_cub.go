package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Longtusk Cub — Creature — Cat {1}{G}, 2/2:
//
//	"Whenever this creature deals combat damage to a player, you get
//	 {E}{E} (two energy counters).
//	 Pay {E}{E}: Put a +1/+1 counter on this creature."
//
// ADR 0129 PR 1 (#1995).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2e78392-096a-4bd0-be19-9c54dfb8451f",
		Name:         "Longtusk Cub",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Longtusk Cub — you get {E}{E}", effects.EbYouGetEnergy(2)),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Pay {E}{E}: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.PayEnergy(2),
			Effect:  effects.PlusOneCountersOnThis(1),
		}},
	})
}
