package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Corpsejack Menace — Creature — Fungus {2}{B}{G}, 4/4 (EDHREC rank
// 1553):
//
//	"If one or more +1/+1 counters would be put on a creature you
//	 control, twice that many +1/+1 counters are put on it instead."
//
// Doubling Season's counter half, narrowed to +1/+1 counters on
// creatures — Hardened Scales' predicate with Doubling Season's
// arithmetic. A CR 614 replacement on the counter event, so it
// composes with the other two through the CR 616 ordering prompt:
// [Scales, Menace] → (N+1)×2, [Menace, Scales] → N×2+1, the affected
// player's choice. The Menace does not double counters put on
// itself as it enters (it has none), and a second Menace doubles
// again, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ca0cc02b-b106-4eca-9388-d4b48dd3be49",
		Name:         "Corpsejack Menace",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventCounterPlaced},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventCounter || ev.CounterName != "+1/+1" || ev.CounterDelta <= 0 {
					return false
				}
				target, ok := g.LookupCardForEffect(ev.CounterTarget)
				return ok && target.IsCreature() && target.Controller == src.Controller
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.CounterDelta *= 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Corpsejack Menace: twice that many +1/+1 counters",
		}},
	})
}
