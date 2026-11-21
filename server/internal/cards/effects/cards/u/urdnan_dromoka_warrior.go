package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Urdnan, Dromoka Warrior — Legendary Creature — Human Warrior {1}{W}, 1/1
// (Foundations Jumpstart, #5929):
//
//	"When Urdnan enters, put a +1/+1 counter on target creature.
//	 Whenever you attack, target attacking creature with a +1/+1 counter on
//	 it gains first strike until end of turn. If that creature has two or
//	 more +1/+1 counters on it, it gains double strike until end of turn
//	 instead."
//
// The entry trigger is Armasaur Guide's counter on any creature (Urdnan
// included). "Whenever you attack" is Bast's shape: once per declaration
// (OncePerBatch), the target chosen as the trigger goes on the stack and
// required to have a +1/+1 counter then (no such creature, no trigger,
// CR 603.3d). The counters are counted again as the ability resolves, so a
// creature that gained or lost counters in between gets the keyword its
// count earns; one that lost them all gets neither (it is no longer a legal
// target, CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "adc1f850-9217-4a90-8e6a-02ed226ef8df",
		Name:         "Urdnan, Dromoka Warrior",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Urdnan — put a +1/+1 counter on target creature",
				effects.PlusOneCounterOnTheTarget),
				effects.TargetCreature("target creature")),
			effects.OncePerBatch(effects.Targeting(
				effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclaredByYou(ev, source.Controller)
				}, "Urdnan — target attacking creature with a +1/+1 counter on it gains first strike until end of turn (double strike instead with two or more)",
					urdnanGrantStrike),
				effects.TargetCreature("target attacking creature with a +1/+1 counter on it", effects.AttackingCreature(), urdnanHasPlusOne()))),
		},
	})
}

// urdnanHasPlusOne is "with a +1/+1 counter on it".
func urdnanHasPlusOne() effects.CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.Counters[game.CounterPlusOne] > 0
	}
}

func urdnanGrantStrike(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		kw := "first strike"
		if c.Counters[game.CounterPlusOne] >= 2 {
			kw = "double strike"
		}
		return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{kw}, Label: "Urdnan — " + kw}.Apply(ctx)
	}
	return nil
}
