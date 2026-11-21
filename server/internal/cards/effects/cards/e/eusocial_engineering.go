package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:         "2f3f5a1c-8b88-4080-bd80-14db4d6c8c50",
		Name:             "Eusocial Engineering",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Warp("{1}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.LandfallPaying("Eusocial Engineering — create a 2/2 colorless Robot artifact creature token (landfall)",
				game.LandfallPayoff{Tokens: 1},
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 colorless Robot artifact"), N: 1})),
		},
	})
}
