package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "d0a4efc3-fa68-48f2-a6d0-2fff93e05c77",
		Name:             "Greenbelt Radical",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Disguise("{5}{G}{G}")},
		Triggered: []game.TriggeredAbility{
			WhenThisIsTurnedFaceUp("Greenbelt Radical — a +1/+1 counter on each creature you control, and they gain trample",
				Do(
					Instruction(func(ctx *Context) error {
						return b11PutCounterOnEachCreatureYouControl(ctx.Game, ctx.Item)
					}),
					GrantKeywordUntilEOT{
						Match:    And(Creature(), YouControl()),
						Keywords: []string{"trample"},
						Label:    "Greenbelt Radical — creatures you control gain trample",
					},
				)),
		},
	})
}
