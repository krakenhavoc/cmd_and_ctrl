package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trusted Advisor — Creature — Human Advisor {U}, 1/2:
//
//	"Your maximum hand size is increased by two.
//	 At the beginning of your upkeep, return a blue creature you
//	 control to its owner's hand."
//
// The increase is a maximum-hand-size static (ADR 0113 §3, #2074),
// folded in CR 613.11 timestamp order (the 2009-10-01 ruling: a Null
// Profusion then a Trusted Advisor is four). The upkeep return is not
// targeted and not optional: the controller picks a blue creature they
// control as it resolves, Trusted Advisor itself included, and with
// none there is nothing to return. Blue is read off the effective
// colours.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "430e4e74-33ee-49f2-aef5-6d09079c5eb7",
		Name:         "Trusted Advisor",
		Completeness: CompletenessFull,
		HandSize:     []game.HandSizeStatic{YourMaxHandSizeChangedBy(2)},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Trusted Advisor — return a blue creature you control", Do(ReturnOneYouControl{
				Match:    func(c game.Card) bool { return c.IsCreature() && c.HasColor("U") },
				Question: "Trusted Advisor — return a blue creature you control to its owner's hand",
			})),
		},
	})
}
