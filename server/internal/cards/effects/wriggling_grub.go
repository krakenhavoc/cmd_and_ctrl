package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d25e8f54-1915-4d0b-b4b1-93ee40da2cfa",
		Name:         "Wriggling Grub",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Wriggling Grub — create two 1/1 black and green Worms",
				Do(CreateToken{Template: TokenCard("1/1 black and green Worm"), N: 2})),
		},
	})
}
