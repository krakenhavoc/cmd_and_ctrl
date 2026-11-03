package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cast_trigger_entry_counters.go — #4686: the catalog side of "that
// creature enters with N additional counters on it", said by a cast
// trigger about the spell it watched, and converge's own entry counter.
// Its own file, so card PRs landing in parallel do not meet in one.

// CountersPerColorSpent is converge on a permanent: "this permanent
// enters with a <kind> counter on it for each color of mana spent to
// cast it" (Wildgrowth Archaic). A CR 614.1c clause read off the
// resolving spell's payment (CastCounts.ColorsSpent), so it composes
// with every other entry clause and Doubling Season sees it. Converge is
// an ability word, not sunburst: an effect that grants sunburst does not
// grant this, and this does not count as sunburst.
func CountersPerColorSpent(kind string) game.EntryCountersFromCast {
	return game.EntryCountersFromCast{
		Kind:  kind,
		Count: func(cast game.CastCounts) int { return cast.ColorsSpent },
	}
}

// WheneverYouCastThatEntersWithCounters is "Whenever you cast a <spell>
// spell, that creature enters with X additional <kind> counters on it",
// where X is `count` of the spell's payment, read as the spell is cast
// and carried on the trigger (Aetherflux Conduit's reason): the payment
// does not change, and a spell gone by resolution has nothing to mark.
//
// On resolution the trigger marks that spell (game.
// MarkSpellEntersWithCountersForEffect), the same mark a spent "next
// creature spell" promise leaves, so the counters ride the creature's
// entry and every replacement sees them. If the trigger is countered or
// the spell has left the stack, nothing is marked and nothing else gets
// the counters; a copy of the spell is a different object and has no
// mark.
//
// `name` is the source's name, for the mark's attribution. The trigger
// goes on the stack even when X is 0, as printed.
func WheneverYouCastThatEntersWithCounters(name string, spell CardPredicate, kind string, count func(game.ManaSpent) int, label string) game.TriggeredAbility {
	t := On(game.EventCast, YouCast(spell), label, func(g *game.Game, item *game.StackItem) error {
		n := item.Params.Amount
		if n <= 0 {
			return nil
		}
		g.MarkSpellEntersWithCountersForEffect(item.Params.Object, game.PromisedCounter{
			Kind:       kind,
			Count:      n,
			Source:     item.SourceCardID,
			SourceName: name,
		})
		return nil
	})
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Object = game.ObjectRef{ID: ev.CardID, Epoch: g.ObjectEpochForEffect(ev.CardID)}
		item.Params.Amount = count(g.StackItemPaidForEffect(ev.CardID).Spent())
		return item
	}
	return t
}

// colorsOfManaSpent is "the number of colors of mana spent to cast it".
func colorsOfManaSpent(m game.ManaSpent) int { return m.ColorCount() }
