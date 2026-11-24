package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kratos, Stoic Father — Legendary Creature — God Warrior {2}{R}{W}, 4/4
// (Secret Lair Drop):
//
//	"Whenever you attack with one or more Gods and whenever a God dies, you
//	 get an experience counter.
//	 At the beginning of your end step, put a number of +1/+1 counters on
//	 target creature equal to the number of experience counters you have.
//	 Partner—Father & son (You can have two commanders if both have this
//	 ability.)"
//
// The first ability is one trigger with two conditions (Aang's experience
// counter, Mordor Trebuchet's attack read). "Attack with one or more Gods" is
// once per declaration (OncePerBatch, however many Gods attack, CR 603.2c) and
// reads the attacker's types through Kratos's text; "a God dies" is one trigger
// per God, whoever controls it, so the batch is keyed on the creature that died
// and a pair of Gods dying together is two experience counters. The end step
// ability targets a creature as it goes on the stack and counts the counters as
// it resolves (CR 608.2c). Partner—Father & son is a deck-construction rule
// read off the oracle text.
//
// No simplification.
func init() {
	const expLabel = "Kratos, Stoic Father — you get an experience counter"
	exp := effects.OncePerBatch(effects.OnAny([]game.EventKind{game.EventAttack, game.EventLTB},
		func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			if ev.Kind == game.EventAttack {
				if !effects.AttackDeclaredByYou(ev, source.Controller) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && c.HasSubtype(source.TextSubtype("God"))
			}
			dead, ok := effects.DiedCreature(ev, g)
			return ok && dead.HasSubtype(source.TextSubtype("God"))
		}, expLabel, effects.YouGetAnExperienceCounter))
	exp.BatchKey = func(ev game.Event, _ *game.Card, _ *game.Game) string {
		if ev.Kind == game.EventLTB {
			return "dies/" + ev.CardID.String()
		}
		return "attacks"
	}
	effects.Register(effects.Spec{
		OracleID:     "c000d5c3-5b0c-4c5c-9919-76640da0a1e4",
		Name:         "Kratos, Stoic Father",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			exp,
			effects.Targeting(effects.AtYourEndStep("Kratos, Stoic Father — put +1/+1 counters on target creature equal to your experience counters",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: effects.ExperienceCounters(g, item.Controller)}.Apply(ctx)
				}),
				effects.TargetCreature("target creature")),
		},
	})
}
