package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Urza's Saga Auras, batch 725 (#5624): the helpers more than one card of the
// batch reads, or that a card needs registered as an on-disk key.

// S82OldB725EnchantedDealtDamageToACreature is "whenever enchanted creature
// deals damage to a creature" (Venomous Fangs): a damage event with a positive
// amount whose source is the creature this Aura is attached to and whose
// recipient is a creature on the battlefield.
func S82OldB725EnchantedDealtDamageToACreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || !source.IsAttachedTo(ev.Source) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	return ok && OnBattlefield(g, ev.Target) && c.IsCreature()
}

// S82OldB725DestroyTheOtherCreature is "destroy the other creature": the
// creature the enchanted one dealt damage to, if it is still on the battlefield.
func S82OldB725DestroyTheOtherCreature(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || !OnBattlefield(g, item.Trigger.Event.Target) {
		return nil
	}
	return DestroyTarget{Target: item.Trigger.Event.Target}.Apply(NewContext(g, item))
}

// S82OldB725PutCounterOnEnchanted is "put a +1/+1 counter on that creature"
// (Vampiric Embrace), read as the enchanted creature: the creature that died
// is gone, so the only creature the clause can mean is the one the Aura
// enchants as the trigger resolves. An Aura that has left puts none.
func S82OldB725PutCounterOnEnchanted(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host := EnchantedCreatureOf(ctx)
	if host == uuid.Nil || !OnBattlefield(g, host) {
		return nil
	}
	return AddCounter{Target: host, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}

// S82OldB725LoseLifeUnlessPaying is Power Taint's trigger body: the player
// whose upkeep it is, the controller of the enchanted permanent, loses N life
// unless they pay cost.
func S82OldB725LoseLifeUnlessPaying(name, cost string, decline game.PayUnlessElse) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil || item.Trigger.Event.Actor == uuid.Nil {
			return nil
		}
		return UpkeepPayUnless{
			Chooser:  item.Trigger.Event.Actor,
			Cost:     cost,
			Question: name + " — pay " + cost + "?",
			Else:     decline,
		}.Apply(NewContext(g, item))
	}
}

// S82OldB725PowerTaintDecline: that player loses 2 life.
var S82OldB725PowerTaintDecline = PayUnlessElse("pay-unless/power-taint-lose-2",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), r.Chooser, -2)
	})

// S82OldB725SacrificeUnlessPayingItsManaCost is Pendrell Flux's granted
// "At the beginning of your upkeep, sacrifice this creature unless you pay its
// mana cost." A creature with no mana cost cannot pay it and is sacrificed.
func S82OldB725SacrificeUnlessPayingItsManaCost(g *game.Game, item *game.StackItem) error {
	if SourceIsNewObject(g, item) || !OnBattlefield(g, item.SourceCardID) {
		return nil
	}
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok {
		return nil
	}
	ctx := NewContext(g, item)
	if c.ManaCost == "" {
		return SacrificePermanent{Target: item.SourceCardID}.Apply(ctx)
	}
	return UpkeepPayUnless{
		Chooser:  item.Controller,
		Cost:     c.ManaCost,
		Question: "Pendrell Flux — pay " + c.ManaCost + ", or sacrifice this creature?",
		Else:     s82OldB725SacrificeThisCreature,
	}.Apply(ctx)
}

var s82OldB725SacrificeThisCreature = PayUnlessBranches("pay-unless/pendrell-flux-sacrifice", nil,
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		if SourceIsNewObject(ctx.Game, ctx.Item) {
			return nil
		}
		return SacrificeThisIfStillOnBattlefield(ctx.Game, ctx.Item)
	})
