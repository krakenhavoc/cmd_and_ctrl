package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bold Plagiarist — Creature — Vampire Rogue {3}{B}, 2/2 (Commander 2021,
// #5504):
//
//	"Flash
//	 Whenever an opponent puts one or more counters on a creature they
//	 control, they put the same number and kind of counters on this
//	 creature."
//
// The engine emits one EventCounterPlaced per permanent and kind, so "one or
// more counters" is the event, the number is its delta and the kind its
// label (Captain Marvel, Apex Avenger's reading). "An opponent puts" is the
// placement's own actor, and the creature is one THAT opponent controls, so
// counters an opponent puts on a creature you control, or on the Plagiarist
// itself, do not trigger it. The copy is mandatory and the counters are put
// BY that opponent (their own replacement effects apply, CR 122.6), not by
// the Plagiarist's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "85277284-1eba-4f87-bf7b-abf1dcbb7691",
		Name:            "Bold Plagiarist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return boldPlagiaristPlaced(ev, source, g)
			}, "Bold Plagiarist — that opponent puts the same number and kind of counters on this creature",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ev := item.Trigger.Event
					n := effects.B33CountersPlacedDelta(ev, ev.Label, g)
					if n <= 0 || !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return g.AddCounterByForEffect(effects.B11CounterPlacerOf(ev, g), item.SourceCardID, ev.Label, n)
				}),
		},
	})
}

// boldPlagiaristPlaced is the trigger's condition: an opponent of the
// Plagiarist's controller put counters on a creature that same opponent
// controls.
func boldPlagiaristPlaced(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventCounterPlaced || ev.Target == source.InstanceID {
		return false
	}
	if effects.B33CountersPlacedDelta(ev, ev.Label, g) <= 0 {
		return false
	}
	placer := effects.B11CounterPlacerOf(ev, g)
	if placer == uuid.Nil || placer == source.Controller {
		return false
	}
	target, ok := g.LookupCardForEffect(ev.Target)
	return ok && effects.OnBattlefield(g, ev.Target) && target.IsCreature() && target.Controller == placer
}
