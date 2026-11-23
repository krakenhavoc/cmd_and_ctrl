package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Conductor of Cacophony — Creature — Demon {3}{B}, 2/1 (Jumpstart 2022,
// #5948):
//
//	"This creature enters with two +1/+1 counters on it.
//	 {B}, Remove a +1/+1 counter from this creature: It deals 1 damage to
//	 each other creature and each player."
//
// The entry counters are Costume Closet's replacement. The counter is
// removed as a cost at announce (CR 602.2b), so a response cannot spend it
// twice. The damage is one instance from the Conductor to every creature
// but itself and every player, the controller included, snapshotted before
// the first point lands. A Conductor that left and came back is a new
// object and is not "other" (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "94daa0c2-4790-4193-aaa8-ba492e8c8fe7",
		Name:         "Conductor of Cacophony",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 2, "Conductor of Cacophony: enters with two +1/+1 counters"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, Remove a +1/+1 counter from this creature: It deals 1 damage to each other creature and each player.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}},
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.RemoveCountersFromThis(game.CounterPlusOne, 1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				self := item.SourceCardID
				if effects.SourceIsNewObject(g, item) {
					self = uuid.Nil
				}
				other := func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID != self }
				return effects.DamageEachMatchingAndEachPlayerThen(ctx, effects.And(effects.Creature(), other), 1, effects.RecipientThen{})
			},
		}},
	})
}
