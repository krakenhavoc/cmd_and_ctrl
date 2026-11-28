package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mountain Titan — Creature — Giant {2}{B}{R}, 2/2 (Ice Age, S82 OLD batch 845, #5636):
//
//	"{1}{R}{R}: Until end of turn, whenever you cast a black spell, put a
//	 +1/+1 counter on this creature."
//
// A repeating delayed triggered ability with a stated duration (CR 603.7b),
// controlled by whoever activated it, one instance for each black spell the
// Titan's controller casts this turn (Bonus Round's shape); activating again
// makes another, and each one puts its own counter. The counter goes on the
// Titan only while it is still the object that made the delayed trigger
// (CR 603.7c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "160c8e94-ccf9-4058-b7fd-04529456f0fa",
		Name:         "Mountain Titan",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}{R}: Until end of turn, whenever you cast a black spell, put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{R}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
					return nil
				}
				return effects.DelayedOnEvent{
					Label:     "Mountain Titan — put a +1/+1 counter on this creature",
					On:        []game.EventKind{game.EventCast},
					Condition: oldB845YouCastABlackSpell,
					Body:      oldB845TitanCounterBody,
					Repeats:   true,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

var (
	// "Whenever you cast a black spell": the trigger's controller cast it
	// (CR 603.7d), and the spell is black as it was cast.
	oldB845YouCastABlackSpell = game.DelayedCondition("cast/you-cast-a-black-spell",
		func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
			if ev.Kind != game.EventCast || ev.Actor != dt.Controller {
				return false
			}
			c, ok := g.LookupCardForEffect(ev.CardID)
			if !ok {
				return false
			}
			for _, col := range c.EffectiveColors() {
				if col == "B" {
					return true
				}
			}
			return false
		})

	// "Put a +1/+1 counter on this creature."
	oldB845TitanCounterBody = game.SimpleDelayedBody("counter/mountain-titan-plus-one", func(g *game.Game, item *game.StackItem) error {
		return effects.PutACounterOnThis(game.CounterPlusOne)(g, item)
	})
)
