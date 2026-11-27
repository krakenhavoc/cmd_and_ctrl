package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

var (
	// "Whenever an attacking creature deals combat damage to you": a combat
	// EventDealDamage to the delayed trigger's controller, from a creature
	// that is attacking and is not the controller's own.
	s82OldB782AttackerHitsYouCondition = game.DelayedCondition("combat/s82-old-b782-attacker-hits-you", s82OldB782AttackerHitsYou)

	// "It deals that much damage to its controller": the damage event's
	// source is the source of the damage, the amount its Amount.
	s82OldB782ThatMuchToItsControllerBody = game.SimpleDelayedBody("damage/s82-old-b782-that-much-to-its-controller", s82OldB782ThatMuchToItsController)
)

func s82OldB782AttackerHitsYou(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || ev.Target != dt.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && c.IsCreature() && c.AttackingTarget != uuid.Nil
}

func s82OldB782ThatMuchToItsController(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ev := item.Trigger.Event
	c, ok := g.LookupCardForEffect(ev.Source)
	if !ok || ev.Amount <= 0 || g.PlayerByIDForEffect(c.Controller) == nil {
		return nil
	}
	return DealDamage{Source: ev.Source, Target: c.Controller, Amount: ev.Amount}.Apply(NewContext(g, item))
}

// S82OldB782HarshJusticeResolve is Harsh Justice: "This turn, whenever an
// attacking creature deals combat damage to you, it deals that much damage to
// its controller." A repeating delayed trigger (CR 603.7b) for the rest of the
// turn, controlled by the spell's controller ("you"), one firing per damage
// event.
func S82OldB782HarshJusticeResolve(_ *game.StackItem, ctx *Context) error {
	return DelayedOnEvent{
		Label:     "Harsh Justice — that creature deals that much damage to its controller",
		On:        []game.EventKind{game.EventDealDamage},
		Condition: s82OldB782AttackerHitsYouCondition,
		Body:      s82OldB782ThatMuchToItsControllerBody,
		Repeats:   true,
	}.Apply(ctx)
}
