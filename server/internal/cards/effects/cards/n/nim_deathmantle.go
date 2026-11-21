package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nim Deathmantle — Artifact — Equipment {2} (Double Masters 2022):
//
//	"Equipped creature gets +2/+2, has intimidate, and is a black Zombie.
//	 Whenever a nontoken creature is put into your graveyard from the
//	 battlefield, you may pay {4}. If you do, return that card to the
//	 battlefield and attach this Equipment to it.
//	 Equip {4}"
//
// "Is a black Zombie" replaces the creature types (CR 205.1a) and the colors
// (CR 105.3) and keeps the card types. The trigger is the controller's own
// graveyard only (the card's owner is the Equipment's controller) and never a
// token. The {4} is paid on resolution; the card must still be that object in
// the graveyard when it is paid, and it comes back under its owner's control
// as a new object, then the Equipment attaches to it, unless the Equipment is
// no longer the permanent that asked (CR 701.3b).
//
// No simplification.
func init() {
	revive := effects.On(game.EventLTB, nimDeathmantleCondition,
		"Nim Deathmantle — you may pay {4} to return that card to the battlefield and attach this Equipment to it",
		nimDeathmantleAsk)
	effects.Register(effects.Spec{
		OracleID:     "66d41377-626d-4ae6-ba86-17bf0c8b3362",
		Name:         "Nim Deathmantle",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("intimidate"),
			effects.SetAttachedCreatureTypes("Zombie"),
			effects.SetAttachedColors("B"),
		},
		Triggered: []game.TriggeredAbility{revive},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{4}")},
	})
}

// nimDeathmantleCondition is "a nontoken creature is put into your
// graveyard from the battlefield": it died as a creature (CR 603.10a), is a
// card, and its owner is the Equipment's controller.
func nimDeathmantleCondition(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := effects.DiedCreature(ev, g)
	return ok && !c.IsToken() && c.Owner == source.Controller
}

func nimDeathmantleAsk(g *game.Game, item *game.StackItem) error {
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{4}",
		Question: "Nim Deathmantle — pay {4} to return that card to the battlefield and attach the Equipment to it?",
		Then:     nimDeathmantlePaid,
	}.Apply(effects.NewContext(g, item))
}

// nimDeathmantlePaid is the paid half, keyed: the card the trigger's event
// names returns from the graveyard it is still in, and the tail attaches.
var nimDeathmantlePaid = effects.PayUnlessBranches("pay-unless/nim-deathmantle-return-and-attach",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		item := ctx.Item
		if item == nil || item.Trigger == nil {
			return nil
		}
		id := item.Trigger.Event.CardID
		z := ctx.Game.FindCardZoneForEffect(id)
		if z == nil || z.Kind != game.ZoneGraveyard {
			return nil
		}
		return ctx.Game.PutOntoBattlefieldTogetherKeyedForEffect(
			[]game.BatchEntry{{CardID: id, From: game.ZoneGraveyard}},
			game.ZoneEntryOptions{Controller: item.Controller},
			game.KeyedTail{Ref: nimDeathmantleAttach, Player: item.Controller, Asked: game.NewAskedItem(item)})
	}, nil)

// nimDeathmantleAttach attaches the asking Equipment to the card that came
// back.
var nimDeathmantleAttach = game.RegisterTail("graveyard/nim-deathmantle-attach", func(g *game.Game, o game.TailOutcome) error {
	if len(o.IDs) == 0 || o.Tail.Asked == nil {
		return nil
	}
	return g.AttachSourceForEffect(o.Tail.Asked.StackItem(), game.TargetRef{Kind: game.TargetCard, ID: o.IDs[0]})
})
