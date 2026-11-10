package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Loading Zone — Enchantment {3}{G} (Edge of Eternities, #4461):
//
//	"If one or more counters would be put on a creature, Spacecraft, or
//	 Planet you control, twice that many of each of those kinds of
//	 counters are put on it instead.
//	 Warp {G} (You may cast this card from your hand for its warp cost.
//	 Exile this enchantment at the beginning of the next end step, then
//	 you may cast it from exile on a later turn.)"
//
// Doubling Season's counter half (a CR 614 replacement on the counter
// event that doubles ev.CounterDelta) narrowed to creatures, Spacecraft
// and Planets you control, read through the layered view as the counters
// would be put. Like Corpsejack Menace and Hardened Scales it names no
// effect, so counters from combat damage, a cost or a turn-based action
// are doubled too; a removal is not a placement (CR 122.6) and is left
// alone. A Spacecraft's or Planet's charge counters from station double,
// as the release notes say. Other doublers order through the CR 616
// prompt. Warp is the shared constructor (Eusocial Engineering).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "16156274-8dc0-439c-94c4-e8c89bbb5687",
		Name:             "Loading Zone",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Warp("{G}")},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventCounterPlaced},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventCounter || ev.CounterDelta <= 0 {
					return false
				}
				target, ok := g.LookupCardForEffect(ev.CounterTarget)
				if !ok || target.Controller != src.Controller {
					return false
				}
				return target.IsCreature() || target.HasSubtype("Spacecraft") || target.HasSubtype("Planet")
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.CounterDelta *= 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Loading Zone: twice that many counters",
		}},
	})
}
