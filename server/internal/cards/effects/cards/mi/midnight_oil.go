package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "26bfc19c-8122-4e95-9a62-2b5418429db2",
		Name:         "Midnight Oil",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(midnightOilHour, 7, "Midnight Oil — enters with seven hour counters"),
		},
		HandSize: []game.HandSizeStatic{{
			Players: game.HandSizeYou,
			Kind:    game.HandSizeSet,
			Dynamic: func(_ *game.Game, source *game.Card) (int, bool) {
				return source.Counters[midnightOilHour], true
			},
		}},
		Triggered: []game.TriggeredAbility{
			// The counters come off after the draw's prompt is answered (a dredge offer;
			// CR 608.2c, #4155).
			effects.On(game.EventBeginDrawStep, effects.ByYou, "Midnight Oil — draw an additional card and remove two hour counters", effects.Do(
				effects.DrawCards{N: 1},
				effects.Instruction(func(ctx *effects.Context) error {
					g := ctx.Game
					c, ok := g.LookupCardForEffect(ctx.Source())
					if !ok || !effects.OnBattlefield(g, ctx.Source()) {
						return nil
					}
					n := min(2, c.Counters[midnightOilHour])
					if n <= 0 {
						return nil
					}
					return effects.AddCounter{Target: ctx.Source(), Kind: midnightOilHour, N: -n}.Apply(ctx)
				}),
			)),
			effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.DiscardedByYou(ev, source)
			}, "Midnight Oil — you lose 1 life", func(g *game.Game, item *game.StackItem) error {
				return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -1)
			}),
		},
	})
}

// midnightOilHour is the hour counter (Midnight Clock's too).
const midnightOilHour = "hour"
