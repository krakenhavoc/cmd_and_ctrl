package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b059_helpers.go — shared pieces of Kaldheim's Saga batch 059
// (#5505): the delayed triggers that last "this turn" on Battle of Frost
// and Fire, Showdown of the Skalds and The Raven's Warning.

var (
	// OldB059YouCastManaValueFiveOrMoreCondition is "whenever you cast a
	// spell with mana value 5 or greater": an EventCast by the delayed
	// trigger's controller of a spell whose mana value on the stack, X
	// counted (CR 202.3), is at least five.
	OldB059YouCastManaValueFiveOrMoreCondition = game.DelayedCondition("old-b059/you-cast-mv-5-or-more",
		func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
			if ev.Kind != game.EventCast || ev.Actor != dt.Controller {
				return false
			}
			spell, ok := g.LookupCardForEffect(ev.CardID)
			if !ok {
				return false
			}
			mv, ok := g.ManaValueForEffect(spell)
			return ok && mv >= 5
		})

	// OldB059DrawTwoThenDiscardOneBody is "draw two cards, then discard a
	// card" (Battle of Frost and Fire's chapter III).
	OldB059DrawTwoThenDiscardOneBody = game.SimpleDelayedBody("old-b059/draw-two-then-discard-one",
		func(g *game.Game, item *game.StackItem) error {
			return DrawThenDiscard(g, item.Controller, item.SourceCardID, 2, 1, "")
		})

	// OldB059FlyingCreatureHitAPlayerCondition is "whenever one or more
	// creatures you control with flying deal combat damage to a player":
	// an each-condition (CR 603.2c) naming the player the damage reached,
	// so with OncePerBatch it fires once for each player a combat damage
	// step's flyers connected with, however many flyers did. The player
	// rides the item as its payload object (ctx.PayloadCards).
	OldB059FlyingCreatureHitAPlayerCondition = game.DelayedEachCondition("old-b059/flying-creatures-hit-a-player",
		func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) []uuid.UUID {
			if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || g.PlayerByIDForEffect(ev.Target) == nil {
				return nil
			}
			src, ok := g.LookupCardForEffect(ev.Source)
			if !ok || !src.IsCreature() || !game.HasKeyword(&src, "flying") || B17DamageSourceController(ev, g) != dt.Controller {
				return nil
			}
			return []uuid.UUID{ev.Target}
		})

	// OldB059LookAtThatHandAndDrawBody is "look at that player's hand and
	// draw a card": the player the damage reached is the payload object.
	OldB059LookAtThatHandAndDrawBody = game.SimpleDelayedBody("old-b059/look-at-that-hand-and-draw",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			for _, id := range ctx.PayloadCards() {
				g.LookAtHandForEffect(ctx.Controller(), id)
			}
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		})
)
