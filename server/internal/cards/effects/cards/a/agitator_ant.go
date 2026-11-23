package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Agitator Ant — Creature — Insect {2}{R}, 2/2:
//
//	"At the beginning of your end step, each player may put two +1/+1
//	 counters on a creature they control. Goad each creature that had
//	 counters put on it this way. (Until your next turn, those creatures
//	 attack each combat if able and attack a player other than you if
//	 able.)"
//
// The fan-out is cmdr2S34EachPlayerMayPutCounters: each player, in turn
// order from the active player, picks one creature they control or none (a
// choice, not a target), and once the last has answered the counters go on
// together. Goad each creature that really got counters: one whose counters
// were prevented, or that left before they went on, is not goaded. The
// goader is the Ant's controller, and the goad ends at their next turn
// (CR 701.15a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "abe377f8-35f0-4dea-a681-91d31a36815c",
		Name:         "Agitator Ant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Agitator Ant — each player may put two +1/+1 counters on a creature they control, then goad each creature that got them",
				func(g *game.Game, item *game.StackItem) error {
					return effects.Cmdr2S34EachPlayerMayPutCounters(effects.NewContext(g, item), effects.Cmdr2S34RiderGoad)
				}),
		},
	})
}
