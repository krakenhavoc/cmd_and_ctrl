package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "430e4e74-33ee-49f2-aef5-6d09079c5eb7",
		Name:         "Trusted Advisor",
		Completeness: effects.CompletenessFull,
		HandSize:     []game.HandSizeStatic{effects.YourMaxHandSizeChangedBy(2)},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Trusted Advisor — return a blue creature you control", effects.Do(effects.ReturnOneYouControl{
				Match:    func(c game.Card) bool { return c.IsCreature() && c.HasColorWord("U") },
				Question: "Trusted Advisor — return a blue creature you control to its owner's hand",
			})),
		},
	})
}
