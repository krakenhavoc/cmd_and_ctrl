package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Norwood Warrior — Creature — Elf Warrior {2}{G}, 2/2 (Portal Second Age, S82 older sets batch 735, #5625):
//
//	"Whenever this creature becomes blocked, it gets +1/+1 until end of turn."
//
// Triggers once as it becomes blocked, however many creatures block it (CR 509.3c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b3ae1481-9fa7-45dd-bbf5-de2d815c4227",
		Name:         "Norwood Warrior",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.Self,
				"Norwood Warrior — it gets +1/+1 until end of turn",
				effects.ThisGetsUntilEndOfTurn(1, 1, "Norwood Warrior — +1/+1")),
		},
	})
}
