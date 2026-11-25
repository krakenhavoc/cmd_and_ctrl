package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sea Gate Stormcaller — Creature — Human Wizard {1}{U}, 2/1 (Zendikar
// Rising, S82 OLD #5508):
//
//	"Kicker {4}{U}
//	 When this creature enters, copy the next instant or sorcery spell with
//	 mana value 2 or less you cast this turn when you cast it. If this
//	 creature was kicked, copy that spell twice instead. You may choose new
//	 targets for the copies."
//
// The enter trigger schedules a delayed trigger (CR 603.7) for the next
// instant or sorcery spell its controller casts this turn whose mana value
// is 2 or less, as it is on the stack (CR 202.3e: an X counts as chosen). A
// cast that does not pass leaves it waiting. The amount is one copy, or two
// from a kicked Stormcaller, and each copy may choose new targets
// (CR 707.10c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "675de007-a71e-4036-8204-140c67871b7f",
		Name:          "Sea Gate Stormcaller",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{4}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sea Gate Stormcaller — copy the next instant or sorcery spell with mana value 2 or less you cast this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					n := 1
					if c, ok := g.LookupCardForEffect(item.SourceCardID); ok && game.CardKickedTimes(c) > 0 {
						n = 2
					}
					return effects.DelayedOnEvent{
						Label:     "Sea Gate Stormcaller — copy that spell",
						On:        []game.EventKind{game.EventCast},
						Condition: seaGateStormcallerCondition,
						Body:      seaGateStormcallerBody,
						Params:    game.EffectParams{Amount: n},
					}.Apply(ctx)
				}),
		},
	})
}

var seaGateStormcallerCondition = game.DelayedCondition("cast/you-next-instant-or-sorcery-mana-value-two-or-less",
	func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
		if ev.Kind != game.EventCast || ev.Actor != dt.Controller {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if !ok || !(c.IsInstant() || c.IsSorcery()) {
			return false
		}
		mv, known := g.CastManaValueForEffect(&c)
		return known && mv <= 2
	})

var seaGateStormcallerBody = game.DelayedBody("copy/sea-gate-stormcaller-the-spell-just-cast",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		ctx := effects.NewContext(g, item)
		cast := ctx.PayloadCards()
		if len(cast) == 0 || p.Amount <= 0 {
			return nil
		}
		return effects.CopySpell{
			StackID:          cast[0],
			Controller:       item.Controller,
			Count:            p.Amount,
			ChooseNewTargets: true,
			FromLastKnown:    true,
		}.Apply(ctx)
	})
