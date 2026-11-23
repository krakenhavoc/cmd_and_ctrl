package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Master of Diversion — Creature — Human Scout {2}{W}, 2/2 (Magic 2014, #5563):
//
//	"Whenever this creature attacks, tap target creature defending player controls."
//
// Distracting Geist's attack trigger: the target is a creature the defending
// player controls, chosen as the trigger goes on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3402aa73-5e63-4158-b114-6f1d42ec8992",
		Name:         "Master of Diversion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.DistractingGeistTap("Master of Diversion — tap target creature defending player controls"),
		},
	})
}
