package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chalice of the Void — Artifact {X}{X}:
//
//	"This artifact enters with X charge counters on it.
//	 Whenever a player casts a spell with mana value equal to the number
//	 of charge counters on this artifact, counter that spell."
//
// Astral Cornucopia's announced X onto charge counters. The trigger is a
// plain cast trigger on every player's spells, Vexing Bauble's shape: the
// condition compares the spell's mana value as it is cast (CR 202.3e
// counts an {X} spell's announced X while it is on the stack) with the
// counters the Chalice has as the spell is cast, and the spell is
// countered as the trigger resolves. A Chalice on zero counters counters
// every spell with mana value 0, and an X spell is read with its X paid.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "74027ba4-a6ba-4696-baf1-60ef0b27ee0a",
		Name:         "Chalice of the Void",
		Completeness: CompletenessFull,
		XMatters:     true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			XCounters(game.CounterCharge),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return source != nil && spellManaValueForEffect(g, ev.CardID) == source.Counters[game.CounterCharge]
			}, "Chalice of the Void — counter that spell", func(g *game.Game, item *game.StackItem) error {
				return CounterTarget{StackID: item.Trigger.Event.CardID}.Apply(NewContext(g, item))
			}),
		},
	})
}
