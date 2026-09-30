package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Plumb the Forbidden — Instant {1}{B}:
//
//	"As an additional cost to cast this spell, you may sacrifice one or
//	 more creatures. When you do, copy this spell for each creature
//	 sacrificed this way.
//	 You draw a card and lose 1 life."
//
// The variable sacrifice (ADR 0100 §3): "you may sacrifice one or more"
// is one clause whose count runs from zero, because sacrificing none is
// not paying — so it is the mandatory SacrificeAnyNumberCost, not an
// optional cost.
//
// "When you do" is a reflexive trigger of the spell (CR 603.12): it
// triggers as the cost is paid, during casting, and goes on the stack
// above the spell once the cast is done — the moment a "when you cast
// this spell" ability goes there, so it is written as one, and it
// fires only when the announcement sacrificed at least one creature.
// The count is read off the payment record as the spell is cast
// (PaidCost.Sacrificed), not at resolution. The copies are created from
// last-known information if the spell has been countered in response
// (CR 608.2h), like storm's: the trigger names the spell and does not
// target it. A copy is not cast (CR 707.12), so a copy never triggers
// this again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "b099fc54-cdbc-46cb-b4e4-7c2ca77b115b",
		Name:           "Plumb the Forbidden",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeAnyNumberCost("one or more creatures", Creature()),
		Triggered: []game.TriggeredAbility{{
			FromStack: true,
			Watches:   []game.EventKind{game.EventCast},
			Key:       "Plumb the Forbidden — copy this spell for each creature sacrificed",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && g.StackItemPaidForEffect(ev.CardID).Sacrificed > 0
			},
			// A fill-in Build (ADR 0041 P9): the caster and the count are
			// facts of the cast; the effect is the row's.
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Plumb the Forbidden — copy this spell for each creature sacrificed")
				item.Controller, item.Owner = ev.Actor, ev.Actor
				item.Params.Amount = g.StackItemPaidForEffect(ev.CardID).Sacrificed
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Params.Amount <= 0 {
					return nil
				}
				return CopySpell{
					StackID:       item.SourceCardID,
					Controller:    item.Controller,
					Count:         item.Params.Amount,
					FromLastKnown: true,
				}.Apply(NewContext(g, item))
			},
		}},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), item.Controller, -1)
		},
	})
}
