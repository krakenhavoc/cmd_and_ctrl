package cardsmi

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Murkfiend Liege — Creature — Horror {2}{G/U}{G/U}{G/U}, 4/4 (Shadowmoor):
//
//	"Other green creatures you control get +1/+1.
//	 Other blue creatures you control get +1/+1.
//	 Untap all green and/or blue creatures you control during each other
//	 player's untap step."
//
// Glen Elendra Liege's two colour anthems (a green-blue creature gets
// +2/+2, and "other" excludes the Liege by instance). The third line is
// not a trigger: it widens the untap step's turn-based action (CR 502.3),
// Seedborn Muse's permission, over the creatures you control that are
// green, blue, or both, the Liege itself included. It is read on each
// other player's untap step only, never on your own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "61d28182-498f-4bbc-bb7a-c5e1ef872dda",
		Name:         "Murkfiend Liege",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16Anthem(effects.G2OtherCreaturesYouControlOfColor("G"), 1, 1),
			effects.B16Anthem(effects.G2OtherCreaturesYouControlOfColor("U"), 1, 1),
		},
		UntapStep: []game.UntapStepPermission{
			effects.UntapDuringEachOtherPlayersUntapStep("Murkfiend Liege — untap all green and/or blue creatures you control",
				func(c game.Card) bool {
					if !c.IsCreature() {
						return false
					}
					colors := c.EffectiveColors()
					return slices.Contains(colors, "G") || slices.Contains(colors, "U")
				}),
		},
	})
}
