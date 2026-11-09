package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	isSea := func() CardPredicate {
		preds := make([]CardPredicate, len(sea))
		for i, t := range sea {
			preds[i] = HasSubtype(t)
		}
		return Or(preds...)
	}
	draw := func(g *game.Game, item *game.StackItem) error {
		return DelayedOnEvent{
			Label:     name + " — draw a card",
			On:        []game.EventKind{game.EventAttack},
			Condition: summonLeviathanSeaCreatureAttacks,
			Body:      drawOneBody,
			Repeats:   true,
		}.Apply(NewContext(g, item))
	}
	Register(Spec{
		OracleID:     "b38d1811-64d7-4a35-a86c-3291eecd8391",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), name+" — ward {2}"),
			ChapterTrigger(1, SagaChapterLabel(name, 1, "return each creature that isn't a Kraken, Leviathan, Merfolk, Octopus, or Serpent to its owner's hand"),
				Do(BounceAllMatching{Match: And(Creature(), Not(isSea()))})),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "until end of turn, whenever a Kraken, Leviathan, Merfolk, Octopus, or Serpent attacks, draw a card"), draw),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "until end of turn, whenever a Kraken, Leviathan, Merfolk, Octopus, or Serpent attacks, draw a card"), draw),
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
