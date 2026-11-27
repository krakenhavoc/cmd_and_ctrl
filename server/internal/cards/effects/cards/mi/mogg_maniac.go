package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mogg Maniac — Creature — Goblin {1}{R}, 1/1 (Stronghold, S82 OLD batch 747,
// #5627):
//
//	"Whenever this creature is dealt damage, it deals that much damage to
//	 target opponent or planeswalker."
//
// Brash Taunter's trigger (any damage, waiting for it to finish so the TOTAL
// dealt together is "that much", CR 120.3, CR 510.2) aimed at an opponent or a
// planeswalker chosen as the trigger goes on the stack. The Maniac is the
// damage source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "53b86704-4f16-498e-8c2c-e661cb85eb9c",
		Name:         "Mogg Maniac",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ThisIsDealtDamageReflected("Mogg Maniac — it deals that much damage to target opponent or planeswalker", false),
		},
	})
}
