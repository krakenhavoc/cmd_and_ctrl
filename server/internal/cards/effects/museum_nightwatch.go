package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Museum Nightwatch — Creature — Centaur Soldier {3}{W}, 3/2 (Murders
// at Karlov Manor, #4469):
//
//	"When this creature dies, create a 2/2 white and blue Detective
//	 creature token.
//	 Disguise {1}{W} (You may cast this card face down for {3} as a 2/2
//	 creature with ward {2}. Turn it face up any time for its disguise
//	 cost.)"
//
// Person of Interest's Detective, made as it dies. A dies trigger looks
// back at the permanent as it last existed (CR 603.10a), so a
// Nightwatch that dies face down, with no text (CR 708.2a), makes
// nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "3891a914-e597-4126-846f-39bf8534eb92",
		Name:             "Museum Nightwatch",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Disguise("{1}{W}")},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Museum Nightwatch — create a 2/2 white and blue Detective creature token",
				Do(CreateToken{Template: TokenCard("2/2 white and blue Detective"), N: 1})),
		},
	})
}
