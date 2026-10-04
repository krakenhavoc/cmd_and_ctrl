package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Midnight Oil — Enchantment {2}{B}{B}:
//
//	"This enchantment enters with seven hour counters on it.
//	 At the beginning of your draw step, draw an additional card and
//	 remove two hour counters from this enchantment.
//	 Your maximum hand size is equal to the number of hour counters on
//	 this enchantment.
//	 Whenever you discard a card, you lose 1 life."
//
// The maximum is a maximum-hand-size static whose number is read from
// the enchantment every time it is asked for (HandSizeStatic.Dynamic,
// ADR 0113 §3, #2074), folded in CR 613.11 timestamp order: a later
// "no maximum" wins over it and it wins over an earlier one (the
// 2016-09-20 ruling). The draw-step trigger always draws; it removes
// two counters, or the one left, or none (the 2016-09-20 rulings).
// The discard trigger fires for every card you discard, the cleanup
// discard included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "26bfc19c-8122-4e95-9a62-2b5418429db2",
		Name:         "Midnight Oil",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(midnightOilHour, 7, "Midnight Oil — enters with seven hour counters"),
		},
		HandSize: []game.HandSizeStatic{{
			Players: game.HandSizeYou,
			Kind:    game.HandSizeSet,
			Dynamic: func(_ *game.Game, source *game.Card) (int, bool) {
				return source.Counters[midnightOilHour], true
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginDrawStep, ByYou, "Midnight Oil — draw an additional card and remove two hour counters", func(g *game.Game, item *game.StackItem) error {
				if err := g.DrawNForEffect(item.Controller, 1); err != nil {
					return err
				}
				c, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok || !onBattlefield(g, item.SourceCardID) {
					return nil
				}
				n := min(2, c.Counters[midnightOilHour])
				if n <= 0 {
					return nil
				}
				return AddCounter{Target: item.SourceCardID, Kind: midnightOilHour, N: -n}.Apply(NewContext(g, item))
			}),
			On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return discardedByYou(ev, source)
			}, "Midnight Oil — you lose 1 life", func(g *game.Game, item *game.StackItem) error {
				return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -1)
			}),
		},
	})
}

// midnightOilHour is the hour counter (Midnight Clock's too).
const midnightOilHour = "hour"
