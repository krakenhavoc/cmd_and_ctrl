package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "e53f5f82-a8e0-44b1-a3be-162a0724755d",
		Name:            "Ambush Krotiq",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ambush Krotiq — return another creature you control to its owner's hand",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					self := ctx.Source()
					return effects.ReturnOneYouControl{
						Match:    func(c game.Card) bool { return c.IsCreature() && c.InstanceID != self },
						Question: "Ambush Krotiq — return another creature you control to its owner's hand",
					}.Apply(ctx)
				}))),
		},
	})
}
