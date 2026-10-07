package effects

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Protean Hydra — Creature — Hydra {X}{G}{G}, 0/0:
//
//	"This creature enters with X +1/+1 counters on it.
//	 If damage would be dealt to this creature, prevent that damage and
//	 remove that many +1/+1 counters from it.
//	 Whenever a +1/+1 counter is removed from this creature, put two
//	 +1/+1 counters on it at the beginning of the next end step."
//
// #2466: the removal trigger is once per counter (CR 603.2c and the
// rulings) — game.TriggeredAbility.PerCounterRemoved — so damage that
// takes three counters off schedules three end-step triggers, each
// putting two counters back. Any removal counts, the prevention clause's
// own included. Each delayed trigger is pinned to the object it was
// scheduled for (CR 400.7): a Hydra that left and came back is a new
// object and gets nothing, and when the Hydra is gone there is nothing
// to put counters on. Counters lost because the Hydra left the
// battlefield were never "removed" (CR 122.2) and trigger nothing.
//
// The prevention half is Magma Pummeler's, without that card's "while it
// has a counter" clause: with no counters the damage is still prevented.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:                   "6e32b958-70d3-4f3d-b10e-d6c8cb27dd93",
		Name:                       "Protean Hydra",
		Completeness:               CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		Replacements: []game.ReplacementEffect{
			PreventDamageDealtTo(PreventionStatic{
				To:    ToThisCreature,
				Then:  removeThatManyCountersFromThisBody,
				Label: "Protean Hydra — prevent damage to it and remove that many +1/+1 counters",
			}),
		},
		Triggered: []game.TriggeredAbility{
			WheneverACounterIsRemovedFromThis(game.CounterPlusOne,
				"Protean Hydra — put two +1/+1 counters on it at the beginning of the next end step",
				func(g *game.Game, item *game.StackItem) error {
					c, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok {
						return nil
					}
					return ScheduleDelayedTrigger{
						Label:  "Protean Hydra — put two +1/+1 counters on it",
						Cards:  []uuid.UUID{item.SourceCardID},
						Body:   proteanHydraTwoCountersBody,
						Params: game.EffectParams{Object: game.ObjectRef{ID: item.SourceCardID, Epoch: c.ObjectEpoch}},
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

var proteanHydraTwoCountersBody = game.DelayedBody("protean-hydra/two-counters", proteanHydraTwoCounters)

// proteanHydraTwoCounters is the delayed trigger's body: two +1/+1
// counters on the Hydra the trigger was scheduled for, if it is still
// that object on the battlefield.
func proteanHydraTwoCounters(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	id := p.Object.ID
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.ObjectEpoch != p.Object.Epoch || !onBattlefield(g, id) {
		return nil
	}
	return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}.Apply(NewContext(g, item))
}
