package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hungering Hydra — {X}{G} Creature — Hydra 0/0 (Commander 2020):
//
//	"This creature enters with X +1/+1 counters on it.
//	 This creature can't be blocked by more than one creature.
//	 Whenever this creature is dealt damage, put that many +1/+1 counters on
//	 it. (It must survive the damage to get the counters.)"
//
// The entry counters are XCounters; the block cap is Battlefront Krushok's
// MaxBlockers on the creature itself. The trigger is Ripjaw Raptor's enrage
// shape with the amount read at the batch end (ThatMuchDamage, CR 603.2c):
// all the damage dealt at once is one trigger. The counters go on only if the
// Hydra is still the same object on the battlefield when the trigger
// resolves, so lethal damage gets none (CR 400.7).
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "d1dde190-8124-4590-9507-801ea4c8cde3",
		Name:                       "Hungering Hydra",
		Completeness:               effects.CompletenessFull,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterPlusOne)},
		XMatters:                   true,
		BlockRules:                 []game.BlockRule{effects.MaxBlockers(effects.OnSelf(), 1)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B35SelfWasDealtDamage(ev, source)
			},
			Key:        "Hungering Hydra — put that many +1/+1 counters on it",
			AtBatchEnd: effects.ThatMuchDamage(),
			Effect:     hungeringHydraGrow,
		}},
	})
}

func hungeringHydraGrow(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.SourceCardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	if o := item.Trigger.Object; o != nil && o.ID == id && g.ObjectEpochForEffect(id) != o.Epoch {
		return nil
	}
	return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: effects.SettledEventAmount(item)}.Apply(effects.NewContext(g, item))
}
