package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Torgal, A Fine Hound — Legendary Creature — Wolf {1}{G}, 2/2 (Final
// Fantasy, #4462):
//
//	"Whenever you cast your first Human creature spell each turn, that
//	 creature enters with an additional +1/+1 counter on it for each Dog
//	 and/or Wolf you control.
//	 {T}: Add one mana of any color."
//
// The trigger is Wildgrowth Archaic's: a cast trigger that, as it
// resolves, marks the spell it watched (MarkSpellEntersWithCountersForEffect),
// so the counters are a CR 614.1c entry clause every replacement sees. Two
// differences. "First … each turn" is read off this turn's casts: the
// spell is a Human creature spell and no earlier spell cast by you this
// turn was one (b25FirstMulticoloredSpellThisTurn's scan, since the
// engine's per-turn tally keeps no per-type count). And the number is a
// multiplier the entry itself counts: the mark carries the subtypes
// (PromisedCounter.PerSubtype, #4462) and the Dogs and Wolves you control
// are counted as the creature enters, not as the trigger resolved, so a
// Dog that arrives, or dies, in between is counted as it is. Torgal
// itself is a Wolf, so it counts. Human, Dog and Wolf are words of
// Torgal's text (CR 612.2).
//
// No simplification.
func init() {
	const label = "Torgal, A Fine Hound — that creature enters with an additional +1/+1 counter on it for each Dog and/or Wolf you control"
	trigger := On(game.EventCast, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		if !YouCast(And(Creature(), OfSubtype("Human")))(ev, source, lki, g) {
			return false
		}
		human := source.TextSubtype("Human")
		for _, prev := range g.EventsThisTurn() {
			if prev.Seq >= ev.Seq {
				break
			}
			if prev.Kind != game.EventCast || prev.Actor != ev.Actor {
				continue
			}
			if c, ok := g.LookupCardForEffect(prev.CardID); ok && c.IsCreature() && c.HasSubtypeWord(human) {
				return false
			}
		}
		return true
	}, label, func(g *game.Game, item *game.StackItem) error {
		g.MarkSpellEntersWithCountersForEffect(item.Params.Object, game.PromisedCounter{
			Kind:       game.CounterPlusOne,
			Count:      1,
			PerSubtype: strings.Join(g.TextReaderForEffect(item).TextSubtypes("Dog", "Wolf"), "|"),
			Source:     item.SourceCardID,
			SourceName: "Torgal, A Fine Hound",
		})
		return nil
	})
	trigger.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Object = game.ObjectRef{ID: ev.CardID, Epoch: g.ObjectEpochForEffect(ev.CardID)}
		return item
	}
	Register(Spec{
		OracleID:     "9d6ab594-aea9-4f13-8802-f3794ee45425",
		Name:         "Torgal, A Fine Hound",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Triggered: []game.TriggeredAbility{trigger},
	})
}
