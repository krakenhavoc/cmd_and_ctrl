package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Summon: Leviathan — Enchantment Creature — Saga Leviathan {4}{U}{U}, 6/6:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Return each creature that isn't a Kraken, Leviathan, Merfolk,
//	 Octopus, or Serpent to its owner's hand.
//	 II, III — Until end of turn, whenever a Kraken, Leviathan, Merfolk,
//	 Octopus, or Serpent attacks, draw a card.
//	 Ward {2}"
//
// Chapter I is the bounce-all of Cresting Mosasaurus over the five named
// creature types, read through the layered view (a changeling is all of
// them; the Saga is itself a Leviathan and stays). Chapters II and III
// schedule a repeating delayed trigger for the turn (CR 603.7b) that fires
// for any player's attacking creature of those types and draws for the
// Saga's controller.
func init() {
	const name = "Summon: Leviathan"
	sea := []string{"Kraken", "Leviathan", "Merfolk", "Octopus", "Serpent"}
	isSea := func() effects.CardPredicate {
		preds := make([]effects.CardPredicate, len(sea))
		for i, t := range sea {
			preds[i] = effects.HasSubtype(t)
		}
		return effects.Or(preds...)
	}
	draw := func(g *game.Game, item *game.StackItem) error {
		return effects.DelayedOnEvent{
			Label:     name + " — draw a card",
			On:        []game.EventKind{game.EventAttack},
			Condition: summonLeviathanSeaCreatureAttacks,
			Body:      effects.DrawOneBody,
			Repeats:   true,
		}.Apply(effects.NewContext(g, item))
	}
	effects.Register(effects.Spec{
		OracleID:     "b38d1811-64d7-4a35-a86c-3291eecd8391",
		Name:         name,
		Completeness: effects.CompletenessFull,
		// Chapter I bounces every creature but the five sea types, as
		// Cresting Mosasaurus does (ADR 0126 owner decision 2).
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepBounce, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), name+" — ward {2}"),
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "return each creature that isn't a Kraken, Leviathan, Merfolk, Octopus, or Serpent to its owner's hand"),
				effects.Do(effects.BounceAllMatching{Match: effects.And(effects.Creature(), effects.Not(isSea()))})),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "until end of turn, whenever a Kraken, Leviathan, Merfolk, Octopus, or Serpent attacks, draw a card"), draw),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "until end of turn, whenever a Kraken, Leviathan, Merfolk, Octopus, or Serpent attacks, draw a card"), draw),
		},
	})
}

var summonLeviathanSeaCreatureAttacks = game.DelayedCondition("attack/a-sea-creature-attacks-summon-leviathan",
	func(ev game.Event, _ *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
		if ev.Kind != game.EventAttack {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if !ok {
			return false
		}
		for _, t := range []string{"Kraken", "Leviathan", "Merfolk", "Octopus", "Serpent"} {
			if c.HasSubtype(t) {
				return true
			}
		}
		return false
	})
