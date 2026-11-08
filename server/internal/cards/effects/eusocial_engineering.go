package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eusocial Engineering — Enchantment {3}{G}{G} (Edge of Eternities,
// #4461):
//
//	"Landfall — Whenever a land you control enters, create a 2/2
//	 colorless Robot artifact creature token.
//	 Warp {1}{G}"
//
// A landfall token maker (Sami, Ship's Engineer's Robot, untapped) with
// warp (CR 702.185): cast from hand for {1}{G}, it is exiled at the
// beginning of the next end step and may be cast from exile on a later
// turn. A land played in between still makes a Robot.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "2f3f5a1c-8b88-4080-bd80-14db4d6c8c50",
		Name:             "Eusocial Engineering",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Warp("{1}{G}")},
		Triggered: []game.TriggeredAbility{
			LandfallPaying("Eusocial Engineering — create a 2/2 colorless Robot artifact creature token (landfall)",
				game.LandfallPayoff{Tokens: 1},
				Do(CreateToken{Template: TokenCard("2/2 colorless Robot artifact"), N: 1})),
		},
	})
}
