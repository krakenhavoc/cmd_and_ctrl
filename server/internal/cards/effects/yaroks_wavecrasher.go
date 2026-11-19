package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yarok's Wavecrasher — Creature — Elemental {3}{U}, 4/4:
//
//	"When this creature enters, return another creature you control to its
//	 owner's hand."
//
// Mandatory (no "may"): with another creature the controller must pick
// one, with none nothing happens (CR 609.3). It is a pick at resolution,
// not a target, and the Wavecrasher itself is never offered.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "30868d69-2ba3-4373-9d47-4bc8978f35d1",
		Name:         "Yarok's Wavecrasher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Yarok's Wavecrasher — return another creature you control to its owner's hand",
				Do(Instruction(func(ctx *Context) error {
					self := ctx.Source()
					return ReturnOneYouControl{
						Match:    func(c game.Card) bool { return c.IsCreature() && c.InstanceID != self },
						Question: "Yarok's Wavecrasher — return another creature you control to its owner's hand",
					}.Apply(ctx)
				}))),
		},
	})
}
