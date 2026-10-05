package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kicker_pair_helpers.go — the bodies the "Kicker [A] and/or [B]"
// leftovers share (#2353): the Apocalypse and Planeshift Battlemages,
// whose two kicker-linked enters abilities (CR 702.33f) differ only in
// what they do.

// targetPlayerDiscards is "target player discards N cards" as a
// trigger's whole Effect — Thunderscape Battlemage's two and Ana
// Battlemage's three. The discarding player picks the cards (CR
// 701.9a), so it opens a real prompt rather than discarding at
// random. A target that left the game is skipped.
func targetPlayerDiscards(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
			return nil
		}
		ctx := NewContext(g, item)
		if len(ctx.LegalTargets()) == 0 {
			return nil
		}
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
			Player: item.Targets[0].ID,
			Source: item.SourceCardID,
			N:      n,
		})
		return nil
	}
}

// castTriggerIfKickedWith is "When you cast this spell, if it was
// kicked with its [cost] kicker, …" (CR 702.33f, CR 603.4) — Wastescape
// Battlemage. The trigger fires from the stack while the Battlemage is
// still a spell, so the kicker record is read off the spell's stack
// item (PaidCost.OptionalCosts) with KickedWithPaid, and the ability's
// controller is the caster (ev.Actor): a spell's Controller field is
// not reliably the caster's (Abby, Cityscape Leveler).
//
// The clause is a cast trigger, not an enters trigger, so it goes on
// the stack above the Battlemage and resolves even if the Battlemage
// is countered.
func castTriggerIfKickedWith(cost, label string, targets *game.TargetSpec, effect Effect) game.TriggeredAbility {
	return game.TriggeredAbility{
		FromStack: true,
		Watches:   []game.EventKind{game.EventCast},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return ev.CardID == source.InstanceID &&
				game.KickedWithPaid(*source, g.StackItemPaidForEffect(ev.CardID).OptionalCosts, cost)
		},
		Targets: targets,
		Key:     label,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Controller, item.Owner = ev.Actor, ev.Actor
			return item
		},
		Effect: effect,
	}
}

// tapTargetThenItDealsItsPowerToItsController is Ana Battlemage's
// "tap target untapped creature and that creature deals damage equal
// to its power to its controller". The creature is the damage source
// (so lifelink, deathtouch and protection see it), its power is read
// after the tap, and its controller is read before the damage.
func tapTargetThenItDealsItsPowerToItsController(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	return DealDamage{Source: id, Target: c.Controller, Amount: c.CurrentPower()}.Apply(ctx)
}

// untappedCreature matches a creature that is not tapped.
func untappedCreature() CardPredicate {
	return And(Creature(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return !c.Tapped })
}
