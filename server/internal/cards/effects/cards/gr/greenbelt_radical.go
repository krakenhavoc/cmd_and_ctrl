package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Greenbelt Radical — Creature — Centaur Citizen {3}{G}, 4/4 (Murders at
// Karlov Manor, #4469):
//
//	"Disguise {5}{G}{G} (You may cast this card face down for {3} as a 2/2
//	 creature with ward {2}. Turn it face up any time for its disguise
//	 cost.)
//	 When this creature is turned face up, put a +1/+1 counter on each
//	 creature you control. Creatures you control gain trample until end
//	 of turn."
//
// The face-up trigger (CR 708.8) puts one counter on each creature its
// controller controls, the Radical included, then grants trample to the
// creatures they control as it resolves (Crash Through's grant), in
// printed order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "d0a4efc3-fa68-48f2-a6d0-2fff93e05c77",
		Name:             "Greenbelt Radical",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Disguise("{5}{G}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisIsTurnedFaceUp("Greenbelt Radical — a +1/+1 counter on each creature you control, and they gain trample",
				effects.Do(
					effects.Instruction(func(ctx *effects.Context) error {
						return effects.B11PutCounterOnEachCreatureYouControl(ctx.Game, ctx.Item)
					}),
					effects.GrantKeywordUntilEOT{
						Match:    effects.And(effects.Creature(), effects.YouControl()),
						Keywords: []string{"trample"},
						Label:    "Greenbelt Radical — creatures you control gain trample",
					},
				)),
		},
	})
}
