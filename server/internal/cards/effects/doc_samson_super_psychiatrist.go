package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Doc Samson, Super Psychiatrist — {4}{G} Legendary Creature — Gamma Doctor
// Hero 3/6 (Marvel Super Heroes, #4455):
//
//	"If you would put one or more counters on a permanent you control, put
//	 that many plus one of each of those kinds of counters on that
//	 permanent instead.
//	 {T}: Add X mana of any one color, where X is Doc Samson's power."
//
// The first ability is Winding Constrictor's permanent half without the
// artifact-or-creature restriction: a CR 614 replacement on the
// counter-placement event, one extra counter of each kind being placed on
// any permanent its controller controls (a land, a planeswalker, a
// Saga). It fires once per kind, which is "of each of those kinds"; two
// Doc Samsons add two. A removal (a negative delta) is not "putting" and
// passes through untouched.
//
// The mana ability is Avatar Kyoshi's shape: one colour picked, X mana of
// it, X read from Doc Samson's power as the ability is activated
// (CR 605.3b), so his own counters count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "699610c9-c991-4faf-b712-d015ef0799ad",
		Name:         "Doc Samson, Super Psychiatrist",
		Completeness: CompletenessFull,
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
				z := g.FindCardZoneForEffect(target.InstanceID)
				return z != nil && z.Kind == game.ZoneBattlefield
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.CounterDelta++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Doc Samson: one more counter of that kind",
		}},
		ManaAbilities: []ManaAbility{{
			Cost:           ManaAbilityCost{Tap: true},
			ProducedAmount: ProducedOneColor(SourcePower),
			Label:          "{T}: Add X mana of any one color, where X is Doc Samson's power",
		}},
	})
}
