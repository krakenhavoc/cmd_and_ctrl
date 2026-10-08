package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gadwick's First Duel — Enchantment — Saga {1}{U}:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Create a Cursed Role token attached to up to one target
//	 creature. (If you control another Role on it, put that one into the
//	 graveyard. Enchanted creature is 1/1.)
//	 II — Scry 2.
//	 III — When you next cast an instant or sorcery spell with mana
//	 value 3 or less this turn, copy that spell. You may choose new
//	 targets for the copy."
//
// Chapter III is a CR 603.7 delayed trigger, the Doublecast shape with a
// mana-value ceiling on the cast it waits for. The ceiling is the
// condition's Amount, read through ManaValueForEffect so an X spell is
// measured at the X it was announced with (CR 202.3e). A cast that does
// not qualify does not use the trigger up (CR 603.7b).
//
// No simplifications.
const gadwickCastKey = "cast/you-next-cast-at-most-mv"

var gadwickNextCastCondition = game.DelayedCondition(gadwickCastKey, youNextCastWithManaValueAtMost)

// youNextCastWithManaValueAtMost is youNextCast with a ceiling on the
// spell's mana value, carried in the params' Amount.
func youNextCastWithManaValueAtMost(ev game.Event, dt *game.DelayedTrigger, g *game.Game, p game.EffectParams) bool {
	if ev.Kind != game.EventCast || ev.Actor != dt.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !p.Filter.Matches(c) {
		return false
	}
	mv, known := g.ManaValueForEffect(c)
	return known && mv <= p.Amount
}

func init() {
	const name = "Gadwick's First Duel"
	Register(Spec{
		OracleID:     "322e3bc1-2dfa-4d5f-848f-a82d9ce02a67",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel(name, 1, "create a Cursed Role token attached to up to one target creature"),
				TargetCreature("up to one target creature").WithCount(0, 1),
				createRoleOnFirstTarget(RoleCursed)),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "scry 2"),
				func(g *game.Game, item *game.StackItem) error {
					return Scry{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
				}),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "when you next cast an instant or sorcery spell with mana value 3 or less this turn, copy it"),
				func(g *game.Game, item *game.StackItem) error {
					return DelayedOnEvent{
						Label:      name + " — copy that spell",
						On:         []game.EventKind{game.EventCast},
						Condition:  gadwickNextCastCondition,
						CondParams: game.EffectParams{Filter: game.CastFilter{Types: []string{"Instant", "Sorcery"}}, Amount: 3},
						Body:       copyTheSpellBody,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
