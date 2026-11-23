package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ezio Auditore da Firenze — Legendary Creature — Human Assassin {1}{B},
// 3/2 (Assassin's Creed, #5948):
//
//	"Menace
//	 Assassin spells you cast have freerunning {B}{B}. (You may cast a
//	 spell for its freerunning cost if you dealt combat damage to a player
//	 this turn with an Assassin or commander.)
//	 Whenever Ezio deals combat damage to a player, you may pay
//	 {W}{U}{B}{R}{G} if that player has 10 or less life. When you do, that
//	 player loses the game."
//
// The static is a granted alternative cost with freerunning's gate
// (FreerunningForSpellsYouCast). The trigger carries the damaged player;
// the life test is read as the trigger resolves, because it is part of the
// option to pay and not an intervening "if", so a player who gained life in
// response is not asked about. Paying queues the CR 603.12 reflexive
// trigger, which makes that player lose the game.
//
// No simplification.
func init() {
	const label = "Ezio Auditore da Firenze — you may pay {W}{U}{B}{R}{G} if that player has 10 or less life; when you do, that player loses the game"
	effects.Register(effects.Spec{
		OracleID:        "3017ac79-37f1-4cc1-8efd-4add027618a2",
		Name:            "Ezio Auditore da Firenze",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{
			effects.FreerunningForSpellsYouCast("{B}{B}", "Assassin"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.CombatDamageToPlayerBy(ev, source.Controller, g) && ev.Source == source.InstanceID
			},
			Key: label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Player = ev.Target
				return item
			},
			Effect: ezioAuditoreOffer,
		}},
	})
}

func ezioAuditoreOffer(g *game.Game, item *game.StackItem) error {
	victim := item.Params.Player
	p := g.PlayerByIDForEffect(victim)
	if p == nil || p.Eliminated || p.Life > 10 {
		return nil
	}
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{W}{U}{B}{R}{G}",
		Question: "Ezio Auditore da Firenze — pay {W}{U}{B}{R}{G}? If you do, " + p.Name + " loses the game.",
		Then:     ezioAuditorePaid,
		Carry:    []uuid.UUID{victim},
	}.Apply(effects.NewContext(g, item))
}

// ezioAuditorePaid is the paid branch, keyed: the reflexive trigger, with
// the damaged player as its payload.
var ezioAuditorePaid = effects.PayUnlessBranches("pay-unless/ezio-auditore-lose",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) == 0 {
			return nil
		}
		return effects.ReflexiveTrigger{
			Label:   "Ezio Auditore da Firenze — that player loses the game",
			Players: []uuid.UUID{r.Carry[0]},
			Body:    ezioAuditoreLoseBody,
		}.Apply(ctx)
	}, nil)

// ezioAuditoreLoseBody makes the payload player lose the game. An on-disk
// key.
var ezioAuditoreLoseBody = game.ReflexiveBody("ezio-auditore/lose-the-game",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		id, ok := effects.GapB37PlayerPayload(effects.NewContext(g, item))
		if !ok {
			return nil
		}
		return effects.LoseTheGame{Player: id}.Apply(effects.NewContext(g, item))
	}), nil)
