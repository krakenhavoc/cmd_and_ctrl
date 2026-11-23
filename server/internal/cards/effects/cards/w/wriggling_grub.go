package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wriggling Grub — Creature — Worm {1}{B}, 1/1 (Foundations Jumpstart,
// #5464):
//
//	"When this creature dies, create two 1/1 black and green Worm
//	 creature tokens."
//
// Pestbrood Sloth's dies trigger with plain Worms (the token table's
// "1/1 black and green Worm"). The tokens are the dead creature's
// controller's (CR 603.7c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d25e8f54-1915-4d0b-b4b1-93ee40da2cfa",
		Name:         "Wriggling Grub",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Wriggling Grub — create two 1/1 black and green Worms",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 black and green Worm"), N: 2})),
		},
	})
}
