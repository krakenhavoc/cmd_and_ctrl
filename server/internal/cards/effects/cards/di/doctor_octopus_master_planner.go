package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Doctor Octopus, Master Planner — Legendary Creature — Human Scientist
// Villain {5}{U}{B}, 4/8:
//
//	"Other Villains you control get +2/+2.
//	 Your maximum hand size is eight.
//	 At the beginning of your end step, if you have fewer than eight
//	 cards in hand, draw cards equal to the difference."
//
// The maximum is a maximum-hand-size static (ADR 0113 §3, #2074),
// folded in CR 613.11 timestamp order (the 2025-09-19 ruling: Spellbook
// then Doctor Octopus is eight, the other order no maximum). The end-step
// draw is AtYourEndStepDrawUpTo, which The Ten Rings shares: an
// intervening-if checked again on resolution, the difference counted
// then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fecfbef-e521-4025-94ea-451c9abde3de",
		Name:         "Doctor Octopus, Master Planner",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Villain"}, Others: true, YoursOnly: true}, 2, 2),
		},
		HandSize: []game.HandSizeStatic{effects.YourMaxHandSizeIs(8)},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStepDrawUpTo("Doctor Octopus, Master Planner", 8),
		},
	})
}
