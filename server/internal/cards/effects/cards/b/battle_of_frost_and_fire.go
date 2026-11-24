package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Battle of Frost and Fire — Enchantment — Saga {3}{U}{R} (Kaldheim, #5505):
//
//	"I — This Saga deals 4 damage to each non-Giant creature and each
//	 planeswalker.
//	 II — Scry 3.
//	 III — Whenever you cast a spell with mana value 5 or greater this turn,
//	 draw two cards, then discard a card."
//
// Chapter I is one damage instance from the Saga to every creature that is
// not a Giant and every planeswalker (Burn Down the House's sweep).
// Chapter III is a repeating delayed trigger (CR 603.7b) that outlives the
// Saga for the rest of the turn and fires on every spell of mana value 5 or
// more its controller casts, X counted on the stack (CR 202.3).
//
// No simplification.
func init() {
	const name = "Battle of Frost and Fire"
	effects.Register(effects.Spec{
		OracleID:     "5a7b89a6-d42d-4fd4-8754-d9c3e3df2118",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{
			Scry:  3,
			Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 4, Partial: true},
		},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "this Saga deals 4 damage to each non-Giant creature and each planeswalker"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.DamageEachMatching(effects.NewContext(g, item),
						effects.Or(effects.And(effects.Creature(), effects.Not(effects.HasSubtype("Giant"))), effects.Planeswalker()), 4)
				}),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "scry 3"), effects.Do(effects.Scry{N: 3})),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "whenever you cast a spell with mana value 5 or greater this turn, draw two cards, then discard a card"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.DelayedOnEvent{
						Label:     name + " — draw two cards, then discard a card",
						On:        []game.EventKind{game.EventCast},
						Condition: effects.OldB059YouCastManaValueFiveOrMoreCondition,
						Body:      effects.OldB059DrawTwoThenDiscardOneBody,
						Repeats:   true,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
