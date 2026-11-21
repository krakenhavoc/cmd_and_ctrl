package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Centaur Battlemaster — Creature — Centaur Warrior {3}{G}{G}, 3/3 (Theros, #5561):
//
//	"Heroic — Whenever you cast a spell that targets this creature, put three +1/+1 counters on this creature."
//
// The shared heroic trigger (CR 207.2c). The counters land only on the same object.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b854749-f3a5-4199-bbfa-83090b8ccc25",
		Name:         "Centaur Battlemaster",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Heroic("Centaur Battlemaster — three +1/+1 counters on it (heroic)", effects.ThsPlusOneCountersOnThisIfSameObject(3)),
		},
	})
}
