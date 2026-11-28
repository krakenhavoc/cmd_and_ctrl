package cardsj

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Justice — Enchantment {2}{W}{W} (Ice Age, S82 OLD batch 840, #5636):
//
//	"At the beginning of your upkeep, sacrifice this enchantment unless you pay
//	 {W}{W}.
//	 Whenever a red creature or spell deals damage, this enchantment deals that
//	 much damage to that creature's or spell's controller."
//
// The upkeep is the Elder Dragons' sacrifice-unless-pay. The second ability is
// Binding Agony's reflection: it triggers on every damage event whose source is a
// red creature (a permanent, or a card that was one as it died) or a red spell, and
// "that much" and the source's controller are recorded as the trigger goes on the
// stack, so a creature that died to its own damage still costs its controller. The
// colour is read through the layered view. A noncreature permanent that is red
// (an enchantment, an artifact) is neither, and does not trigger it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0586fdc7-3025-4040-8eaf-f8a57a21aff0",
		Name:         "Justice",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.UpkeepSacrificeUnlessYouPay("Justice", "{W}{W}"),
			{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
					_, ok := justiceSource(ev, g)
					return ok
				},
				Key: "Justice — that much damage to that creature's or spell's controller",
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, "Justice — that much damage to that creature's or spell's controller")
					item.Params.Amount = ev.Amount
					item.Params.Player, _ = justiceSource(ev, g)
					return item
				},
				Effect: effects.S82OldB811DealRecordedDamageToRecordedPlayer,
			},
		},
	})
}

// justiceSource is the controller of the red creature or red spell that dealt the
// damage the event records, and whether there is one. A spell is one while it is on
// the stack or resolving; a creature is a permanent on the battlefield, or a card
// that was a creature as it left (its card is a creature card).
func justiceSource(ev game.Event, g *game.Game) (uuid.UUID, bool) {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Source == uuid.Nil {
		return uuid.Nil, false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	if !ok || !slices.Contains(c.EffectiveColors(), "R") {
		return uuid.Nil, false
	}
	if controller, isSpell := g.DamageSourceSpellControllerForEffect(ev.Source); isSpell {
		return controller, true
	}
	if !c.IsCreature() {
		return uuid.Nil, false
	}
	if ev.Actor != uuid.Nil {
		return ev.Actor, true
	}
	return c.Controller, true
}
