package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ambush Krotiq — Creature — Insect {5}{G}, 5/5 (Fate Reforged,
// #5552):
//
//	"Trample
//	 When this creature enters, return another creature you control to
//	 its owner's hand."
//
// Trample is printed. The entry trigger is Yarok's Wavecrasher's: mandatory,
// a pick at resolution and not a target, the Krotiq itself never offered;
// with no other creature nothing happens (CR 609.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e53f5f82-a8e0-44b1-a3be-162a0724755d",
		Name:            "Ambush Krotiq",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ambush Krotiq — return another creature you control to its owner's hand",
				Do(Instruction(func(ctx *Context) error {
					self := ctx.Source()
					return ReturnOneYouControl{
						Match:    func(c game.Card) bool { return c.IsCreature() && c.InstanceID != self },
						Question: "Ambush Krotiq — return another creature you control to its owner's hand",
					}.Apply(ctx)
				}))),
		},
	})
}
